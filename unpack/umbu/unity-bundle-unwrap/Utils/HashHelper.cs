namespace Ankama.Localization.Utils
{
    /// <summary>
    /// Provides helper methods for hashing operations.
    /// </summary>
    public static class HashHelper
    {
        // FNV-1a constants
        private const ulong FnvOffsetBasis = 14695981039346656037UL;
        private const ulong FnvPrime = 1099511628211UL;

        /// <summary>
        /// Computes the FNV-1a hash for the given input string.
        /// </summary>
        /// <param name="input">The input string to hash.</param>
        /// <returns>The computed 64-bit hash value.</returns>
        public static ulong Fnv1A(string input)
        {
            if (input == null)
                throw new ArgumentNullException(nameof(input));

            ulong hash = FnvOffsetBasis;
            foreach (char c in input)
            {
                hash ^= (byte)c;
                hash *= FnvPrime;
            }

            return hash;
        }
    }

    /// <summary>
    /// Represents a 64-bit hash value.
    /// </summary>
    public struct Hash64
    {
        private ulong _rawValue;

        /// <summary>
        /// Initializes a new instance of the <see cref="Hash64"/> struct with a hash generated from the specified input string.
        /// </summary>
        /// <param name="input">The input string to hash.</param>
        public Hash64(string input)
        {
            if (input == null)
                throw new ArgumentNullException(nameof(input));

            _rawValue = HashHelper.Fnv1A(input);
        }

        /// <summary>
        /// Creates a new <see cref="Hash64"/> from a raw hash value.
        /// </summary>
        /// <param name="rawHashValue">The raw hash value.</param>
        /// <returns>A <see cref="Hash64"/> initialized with the specified value.</returns>
        public static Hash64 FromRawHashedValue(ulong rawHashValue)
        {
            return new Hash64 { _rawValue = rawHashValue };
        }

        /// <summary>
        /// Gets or sets the raw 64-bit hash value.
        /// </summary>
        public ulong RawValue
        {
            get => _rawValue;
            set => _rawValue = value;
        }
    }
}
