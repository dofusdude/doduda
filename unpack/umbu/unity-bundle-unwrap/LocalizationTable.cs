using System.Text;
using Ankama.Localization.Utils;

namespace Ankama.Localization
{
    /// <summary>
    /// Represents a localization table for managing localizable strings.
    /// </summary>
    public class LocalizationTable
    {
        private readonly LocalizationTableHeader _header;
        private readonly FileStream _stream;
        private readonly BinaryReader _reader;

        public LocalizationTableHeader Header => _header;

        /// <summary>
        /// Initializes a new instance of the <see cref="LocalizationTable"/> class.
        /// </summary>
        /// <param name="header">The header containing metadata about the table.</param>
        /// <param name="stream">The file stream for the table.</param>
        /// <param name="reader">The binary reader for the table.</param>
        public LocalizationTable(LocalizationTableHeader header, FileStream stream, BinaryReader reader)
        {
            _header = header ?? throw new ArgumentNullException(nameof(header));
            _stream = stream ?? throw new ArgumentNullException(nameof(stream));
            _reader = reader ?? throw new ArgumentNullException(nameof(reader));
        }

        /// <summary>
        /// Reads a localization table from a file path.
        /// </summary>
        /// <param name="path">The file path.</param>
        /// <returns>A new instance of <see cref="LocalizationTable"/>.</returns>
        public static LocalizationTable ReadFrom(string path)
        {
            if (string.IsNullOrEmpty(path))
                throw new ArgumentNullException(nameof(path));

            using var stream = new FileStream(path, FileMode.Open, FileAccess.Read);
            using var reader = new BinaryReader(stream);
            var header = LocalizationTableHeader.ReadFrom(reader);
            return new LocalizationTable(header, stream, reader);
        }

        /// <summary>
        /// Attempts to lookup a string value by its key.
        /// </summary>
        /// <param name="key">The key to look up.</param>
        /// <param name="output">The resulting string, if found.</param>
        /// <returns><c>true</c> if the key exists; otherwise, <c>false</c>.</returns>
        public bool TryLookup(string key, out string output)
        {
            if (string.IsNullOrEmpty(key))
            {
                output = null;
                return false;
            }

            if (_header.TryGetOffset(key, out var offset))
            {
                output = ReadStringAtOffset(offset);
                return true;
            }

            output = null;
            return false;
        }

        /// <summary>
        /// Attempts to lookup a string value by its integer key.
        /// </summary>
        /// <param name="key">The integer key to look up.</param>
        /// <param name="output">The resulting string, if found.</param>
        /// <returns><c>true</c> if the key exists; otherwise, <c>false</c>.</returns>
        public bool TryLookup(int key, out string output)
        {
            if (_header.TryGetOffset(key, out var offset))
            {
                output = ReadStringAtOffset(offset);
                return true;
            }

            output = null;
            return false;
        }

        /// <summary>
        /// Reads a string from the binary file at the given offset.
        /// </summary>
        /// <param name="offset">The offset to read from.</param>
        /// <returns>The string at the given offset.</returns>
        private string ReadStringAtOffset(uint offset)
        {
            // Seek to the offset
            _stream.Seek(offset, SeekOrigin.Begin);

            // Read the length of the string
    int length = DecodeCustomLength();

            if (length == 0)
            {
                return string.Empty;
            }

            // Read the bytes of the string
            byte[] bytes = _reader.ReadBytes(length);

            // Convert bytes to string using UTF-8
            string decodedString = Encoding.UTF8.GetString(bytes);

            // Replace non-printable characters with their intended printable representation
            decodedString = EscapeAndHandleUnicodeCharacters(decodedString);

            return decodedString;
        }

        private int DecodeCustomLength()
        {
            byte firstByte = _reader.ReadByte();

            if ((firstByte & 0x80) != 0)
            {
                // VarInt decoding
                int length = firstByte & 0x7F;
                int shift = 7;

                while (true)
                {
                    byte nextByte = _reader.ReadByte();
                    length |= (nextByte & 0x7F) << shift;

                    if ((nextByte & 0x80) == 0)
                        break;

                    shift += 7;
                }

                return length;
            }
            else
            {
                // Single-byte length
                return firstByte;
            }
        }

        private string EscapeAndHandleUnicodeCharacters(string input)
        {
            var sb = new StringBuilder();
            foreach (char c in input)
            {

                if (c == '\uFEFF') // Handle BOM specifically
                {
                }
                else if (char.IsControl(c) && c != '\n' && c != '\r') // Handle control characters but allow newlines and carriage returns
                {
                }
                else if (char.IsWhiteSpace(c) && c == '\u00A0') // Specific handling for non-breaking space
                {
                    sb.Append(" ");
                }
                else
                {
                    sb.Append(c); // Preserve other characters as-is
                }
            }
            return sb.ToString();
        }


        /// <summary>
        /// Writes the localization table to a file.
        /// </summary>
        /// <param name="path">The file path to write to.</param>
        /// <param name="languageCode">The language code.</param>
        /// <param name="integerKeyedStrings">The dictionary of integer-keyed strings.</param>
        /// <param name="stringKeyedStrings">The dictionary of string-keyed strings.</param>
        public static void Write(
            string path,
            string languageCode,
            Dictionary<int, string> integerKeyedStrings,
            Dictionary<string, string> stringKeyedStrings)
        {
            if (string.IsNullOrEmpty(path))
                throw new ArgumentNullException(nameof(path));

            if (string.IsNullOrEmpty(languageCode))
                throw new ArgumentNullException(nameof(languageCode));

            if (integerKeyedStrings == null || stringKeyedStrings == null)
                throw new ArgumentNullException("Dictionaries cannot be null.");

            using var stream = new FileStream(path, FileMode.Create, FileAccess.Write);
            using var writer = new BinaryWriter(stream);

            // Serialization logic (to be implemented)
        }

        /// <summary>
        /// Writes the internal structure of the localization table.
        /// </summary>
        /// <param name="path">The file path.</param>
        /// <param name="languageCode">The language code.</param>
        /// <param name="integerKeyedOffsets">The integer key offsets.</param>
        /// <param name="stringKeyedOffsets">The string key offsets.</param>
        /// <param name="rawStrings">The raw strings stream.</param>
        public static void WriteInternal(
            string path,
            string languageCode,
            Dictionary<int, uint> integerKeyedOffsets,
            Dictionary<Hash64, uint> stringKeyedOffsets,
            Stream rawStrings)
        {
            if (string.IsNullOrEmpty(path))
                throw new ArgumentNullException(nameof(path));

            if (string.IsNullOrEmpty(languageCode))
                throw new ArgumentNullException(nameof(languageCode));

            if (integerKeyedOffsets == null || stringKeyedOffsets == null || rawStrings == null)
                throw new ArgumentNullException("Parameters cannot be null.");

            // Internal writing logic (to be implemented)
        }
    }
}
