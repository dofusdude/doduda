using DotMake.CommandLine;
using AssetsTools.NET;
using AssetsTools.NET.Extra;
using Newtonsoft.Json.Linq;
using Ankama.Localization;
using Ankama.Localization.Utils;
using System.Text.Json;
using System.Collections.Generic;

Cli.Run<DodudaBundleUnpack>(args);

[CliCommand(Description = "The root cli command")]
public class DodudaBundleUnpack
{
    [CliArgument(Description = "File path to the input bundle")]
    public required string InBundlePath { get; set; }
 
    [CliArgument(Description = "Output json file path")]
    public required string OutJsonPath { get; set; }

    // Full Credit to UABEA (MIT License) https://github.com/nesrak1/UABEA/blob/master/UABEAvalonia/Logic/AssetImportExport.cs#L186C24-L186C39
    private static JToken RecurseJsonDump(AssetTypeValueField field, bool uabeFlavor)
    {
        AssetTypeTemplateField template = field.TemplateField;

        bool isArray = template.IsArray;

        if (isArray)
        {
            JArray jArray = new JArray();

            if (template.ValueType != AssetValueType.ByteArray)
            {
                for (int i = 0; i < field.Children.Count; i++)
                {
                    jArray.Add(RecurseJsonDump(field.Children[i], uabeFlavor));
                }
            }
            else
            {
                byte[] byteArrayData = field.AsByteArray;
                for (int i = 0; i < byteArrayData.Length; i++)
                {
                    jArray.Add(byteArrayData[i]);
                }
            }

            return jArray;
        }
        else
        {
            if (field.Value != null)
            {
                AssetValueType evt = field.Value.ValueType;
                
                if (field.Value.ValueType != AssetValueType.ManagedReferencesRegistry)
                {
                    object value = evt switch
                    {
                        AssetValueType.Bool => field.AsBool,
                        AssetValueType.Int8 or
                        AssetValueType.Int16 or
                        AssetValueType.Int32 => field.AsInt,
                        AssetValueType.Int64 => field.AsLong,
                        AssetValueType.UInt8 or
                        AssetValueType.UInt16 or
                        AssetValueType.UInt32 => field.AsUInt,
                        AssetValueType.UInt64 => field.AsULong,
                        AssetValueType.String => field.AsString,
                        AssetValueType.Float => field.AsFloat,
                        AssetValueType.Double => field.AsDouble,
                        _ => "invalid value"
                    };

                    return (JValue)JToken.FromObject(value);
                }
                else
                {
                    // todo separate method
                    ManagedReferencesRegistry registry = field.Value.AsManagedReferencesRegistry;

                    if (registry.version == 1 || registry.version == 2)
                    {
                        JArray jArrayRefs = new JArray();

                        foreach (AssetTypeReferencedObject refObj in registry.references)
                        {
                            AssetTypeReference typeRef = refObj.type;

                            JObject jObjManagedType = new JObject
                            {
                                { "class", typeRef.ClassName },
                                { "ns", typeRef.Namespace },
                                { "asm", typeRef.AsmName }
                            };

                            JObject jObjData = new JObject();

                            foreach (AssetTypeValueField child in refObj.data)
                            {
                                jObjData.Add(child.FieldName, RecurseJsonDump(child, uabeFlavor));
                            }

                            JObject jObjRefObject;

                            if (registry.version == 1)
                            {
                                jObjRefObject = new JObject
                                {
                                    { "type", jObjManagedType },
                                    { "data", jObjData }
                                };
                            }
                            else
                            {
                                jObjRefObject = new JObject
                                {
                                    { "rid", refObj.rid },
                                    { "type", jObjManagedType },
                                    { "data", jObjData }
                                };
                            }

                            jArrayRefs.Add(jObjRefObject);
                        }

                        JObject jObjReferences = new JObject
                        {
                            { "version", registry.version },
                            { "RefIds", jArrayRefs }
                        };

                        return jObjReferences;
                    }
                    else
                    {
                        throw new NotSupportedException($"Registry version {registry.version} not supported!");
                    }
                }
            }
            else
            {
                JObject jObject = new JObject();

                foreach (AssetTypeValueField child in field)
                {
                    jObject.Add(child.FieldName, RecurseJsonDump(child, uabeFlavor));
                }

                return jObject;
            }
        }
    }

    /// <summary>
    /// Dumps all localization values from the accessor into a dictionary.
    /// </summary>
    /// <param name="accessor">The LocalizationAccessor instance.</param>
    /// <returns>A dictionary containing all localization keys and their corresponding values.</returns>
    static Dictionary<string, Dictionary<string, string>> DumpAllValues(LocalizationAccessor accessor)
    {
        Dictionary<string, Dictionary<string, string>> result = new Dictionary<string, Dictionary<string, string>>();
        result["entries"] = new Dictionary<string, string>();

        Console.WriteLine($"Dumping: {accessor.Table.Header.IntegerKeyedOffsets.Count()} Int Values");
        foreach (var key in accessor.Table.Header.IntegerKeyedOffsets.Keys)
        {
            if (accessor.TryGetLocalization(key, out string value))
            {
                result["entries"][key.ToString()] = value;
            }
        }

        Console.WriteLine($"Dumping: {accessor.Table.Header.StringKeyedOffsets.Count()} String Values");
        // Dump string-keyed values
        foreach (var keyHash in accessor.Table.Header.StringKeyedOffsets.Keys)
        {
            string key = keyHash.ToString(); // Assuming Hash64 has a valid ToString implementation
            if (accessor.TryGetLocalization(key, out string value))
            {
                result["entries"][key] = value;
            }
        }

        return result;
    }

    static AssetTypeValueField? LoadBundleFile(string filePath)
    {
        var manager = new AssetsManager();

        var bunInst = manager.LoadBundleFile(filePath, true);
        var afileInst = manager.LoadAssetsFileFromBundle(bunInst, 0, false);
        var afile = afileInst.file;

        // make sure only one MonoBehaviour is in the file
        var a = afile.GetAssetsOfType(AssetClassID.MonoBehaviour);
        if (a.Count != 1)
        {
            Console.WriteLine($"Expected 1 MonoBehaviour, found {a.Count}");
            return null;
        }

        return manager.GetBaseField(afileInst, a[0]);
    }

    public void Run()
    {
        if (!File.Exists(InBundlePath))
        {
            Console.WriteLine($"Input file {InBundlePath} does not exist.");
            return;
        }

        // get absolute path to the output directory
        var outDir = Path.GetDirectoryName(OutJsonPath);
        if (outDir == null || outDir == "") {
            outDir = Directory.GetCurrentDirectory();
        }
        outDir = Path.GetFullPath(outDir);

        // check if input file is ".bin"
        if (Path.GetExtension(InBundlePath) == ".bin")
        {
            // Open the binary file
            using (FileStream fs = new FileStream(InBundlePath, FileMode.Open, FileAccess.Read))
            using (BinaryReader reader = new BinaryReader(fs))
            {
                // Read the LocalizationTableHeader from the file
                LocalizationTableHeader header = LocalizationTableHeader.ReadFrom(reader);

                // Load the LocalizationTable
                LocalizationTable table = new LocalizationTable(header, fs, reader);

                // Create the LocalizationAccessor
                LocalizationAccessor accessor = new LocalizationAccessor(table);

                // Dump all values to JSON
                var allValues = DumpAllValues(accessor);

                // Write the JSON to the output file
                var options = new JsonSerializerOptions
                {
                    Encoder = System.Text.Encodings.Web.JavaScriptEncoder.UnsafeRelaxedJsonEscaping, // Allow unescaped Unicode
                    WriteIndented = true // For pretty-printing
                };
                File.WriteAllText(OutJsonPath, JsonSerializer.Serialize(allValues, options));
            }
        } else {
            var bundle = LoadBundleFile(InBundlePath);
            if (bundle == null)
            {
                Console.WriteLine("Failed to load bundle file.");
                return;
            }

            var json = RecurseJsonDump(bundle, false);

            File.WriteAllText(OutJsonPath, json.ToString());
        }

    }
}