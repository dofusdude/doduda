using Ankama.Localization.Utils;

namespace Ankama.Localization
{
    /// <summary>
    /// Represents the header of a localization table.
    /// </summary>
    public class LocalizationTableHeader
    {
        private readonly Dictionary<int, uint> _integerKeyedOffsets;
        private readonly Dictionary<Hash64, uint> _stringKeyedOffsets;
        private readonly string _languageCode;

        public Dictionary<int, uint> IntegerKeyedOffsets => _integerKeyedOffsets;
        public Dictionary<Hash64, uint> StringKeyedOffsets => _stringKeyedOffsets;

        /// <summary>
        /// Initializes a new instance of the <see cref="LocalizationTableHeader"/> class.
        /// </summary>
        /// <param name="languageCode">The language code for this localization table.</param>
        /// <param name="integerKeyedOffsets">Offsets for integer keys.</param>
        /// <param name="stringKeyedOffsets">Offsets for string keys (hashed).</param>
        public LocalizationTableHeader(string languageCode, Dictionary<int, uint> integerKeyedOffsets, Dictionary<Hash64, uint> stringKeyedOffsets)
        {
            _languageCode = languageCode ?? throw new ArgumentNullException(nameof(languageCode));
            _integerKeyedOffsets = integerKeyedOffsets ?? throw new ArgumentNullException(nameof(integerKeyedOffsets));
            _stringKeyedOffsets = stringKeyedOffsets ?? throw new ArgumentNullException(nameof(stringKeyedOffsets));
        }

        /// <summary>
        /// Gets the language code of the localization table.
        /// </summary>
        public string LanguageCode => _languageCode;

        /// <summary>
        /// Converts a string to its corresponding hash value.
        /// </summary>
        /// <param name="input">The string to hash.</param>
        /// <returns>The hash value.</returns>
        public static Hash64 StringToHash(string input)
        {
            if (string.IsNullOrEmpty(input))
                throw new ArgumentNullException(nameof(input));

            return new Hash64(input); // Assuming `Hash64` struct has a constructor for this.
        }

        /// <summary>
        /// Tries to get the offset for a given integer key.
        /// </summary>
        /// <param name="key">The integer key.</param>
        /// <param name="offset">The offset value, if found.</param>
        /// <returns><c>true</c> if the offset was found; otherwise, <c>false</c>.</returns>
        public bool TryGetOffset(int key, out uint offset)
        {
            return _integerKeyedOffsets.TryGetValue(key, out offset);
        }

        /// <summary>
        /// Tries to get the offset for a given string key (hashed).
        /// </summary>
        /// <param name="key">The string key to hash and lookup.</param>
        /// <param name="offset">The offset value, if found.</param>
        /// <returns><c>true</c> if the offset was found; otherwise, <c>false</c>.</returns>
        public bool TryGetOffset(string key, out uint offset)
        {
            if (string.IsNullOrEmpty(key))
            {
                offset = 0;
                return false;
            }

            var hash = StringToHash(key);
            return _stringKeyedOffsets.TryGetValue(hash, out offset);
        }

        /// <summary>
        /// Reads a localization table header from a binary reader.
        /// </summary>
        /// <param name="reader">The binary reader.</param>
        /// <returns>A new <see cref="LocalizationTableHeader"/> instance.</returns>
        public static LocalizationTableHeader ReadFrom(BinaryReader reader)
        {
            if (reader == null)
                throw new ArgumentNullException(nameof(reader));

            byte version = reader.ReadByte();
            string languageCode = $"{(char)reader.ReadByte()}{(char)reader.ReadByte()}";
            Console.WriteLine($"Localization table version: {version}, language code: {languageCode}");
            Console.WriteLine($"Reading Int Keyed Table");
            var integerKeyedOffsets = ReadIntegerKeyedTable(reader);
            //Console.WriteLine($"Reading String Keyed Table");
            //var stringKeyedOffsets = ReadStringKeyedTable(reader);

            return new LocalizationTableHeader(languageCode, integerKeyedOffsets, new Dictionary<Hash64, uint>());
        }

        /// <summary>
        /// Writes the header to a binary writer.
        /// </summary>
        /// <param name="writer">The binary writer.</param>
        /// <param name="offset">The starting offset.</param>
        public void Write(BinaryWriter writer, uint offset)
        {
            if (writer == null)
                throw new ArgumentNullException(nameof(writer));

            writer.Write(_languageCode);
            WriteIntegerKeyedTable(writer, _integerKeyedOffsets);
            WriteStringKeyedTable(writer, _stringKeyedOffsets);
        }

        /// <summary>
        /// Reads a table of integer-keyed offsets from the binary reader.
        /// </summary>
        /// <param name="reader">The binary reader.</param>
        /// <returns>A dictionary of integer-keyed offsets.</returns>
        public static Dictionary<int, uint> ReadIntegerKeyedTable(BinaryReader reader)
        {
            int count = reader.ReadInt32();
            var table = new Dictionary<int, uint>(count);
            Console.WriteLine($"Reading {count} integer keyed offsets");
            for (int i = 0; i < count; i++)
            {
                int key = reader.ReadInt32();
                uint offset = reader.ReadUInt32();
                table[key] = offset;
            }

            return table;
        }

        /// <summary>
        /// Reads a table of string-keyed offsets from the binary reader.
        /// </summary>
        /// <param name="reader">The binary reader.</param>
        /// <returns>A dictionary of hashed string-keyed offsets.</returns>
        public static Dictionary<Hash64, uint> ReadStringKeyedTable(BinaryReader reader)
        {
            int count = reader.ReadInt32();
            var table = new Dictionary<Hash64, uint>(count);

            for (int i = 0; i < count; i++)
            {
                string keyString = reader.ReadString(); // Read the original string representation
                var hash = new Hash64(keyString);      // Create Hash64 using the string constructor
                uint offset = reader.ReadUInt32();
                table[hash] = offset;
            }

            return table;
        }

        /// <summary>
        /// Writes a table of integer-keyed offsets to the binary writer.
        /// </summary>
        /// <param name="writer">The binary writer.</param>
        /// <param name="table">The table to write.</param>
        private static void WriteIntegerKeyedTable(BinaryWriter writer, Dictionary<int, uint> table)
        {
            writer.Write(table.Count);

            foreach (var pair in table)
            {
                writer.Write(pair.Key);
                writer.Write(pair.Value);
            }
        }

        /// <summary>
        /// Writes a table of string-keyed offsets to the binary writer.
        /// </summary>
        /// <param name="writer">The binary writer.</param>
        /// <param name="table">The table to write.</param>
        private static void WriteStringKeyedTable(BinaryWriter writer, Dictionary<Hash64, uint> table)
        {
            writer.Write(table.Count);

            foreach (var pair in table)
            {
                writer.Write(pair.Key.RawValue); // Assuming `Hash64` exposes a `RawValue` property.
                writer.Write(pair.Value);
            }
        }
    }
}
