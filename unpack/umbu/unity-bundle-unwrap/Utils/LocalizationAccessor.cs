namespace Ankama.Localization.Utils
{
    public class LocalizationAccessor
    {
        private readonly LRUCache<int, string> _integerKeyCache;
        private readonly LRUCache<string, string> _stringKeyCache;
        private readonly LocalizationTable _table;

        public LocalizationTable Table => _table;

        public LocalizationAccessor(LocalizationTable table, int stringCacheSize = 512, int integerCacheSize = 512)
        {
            _table = table ?? throw new ArgumentNullException(nameof(table));
            _integerKeyCache = new LRUCache<int, string>(integerCacheSize);
            _stringKeyCache = new LRUCache<string, string>(stringCacheSize);
        }

        public bool TryGetLocalization(int key, out string localization)
        {
            // Check the integer key cache
            if (_integerKeyCache.TryGetValue(key, out localization))
            {
                return true;
            }

            // Fallback to the table lookup
            if (_table.TryLookup(key, out localization))
            {
                _integerKeyCache.TryAdd(key, localization);
                return true;
            }

            return false;
        }

        public bool TryGetLocalization(string key, out string localization)
        {
            // Check the string key cache
            if (_stringKeyCache.TryGetValue(key, out localization))
            {
                return true;
            }

            // Fallback to the table lookup
            if (_table.TryLookup(key, out localization))
            {
                _stringKeyCache.TryAdd(key, localization);
                return true;
            }

            return false;
        }

        public void CloseTable()
        {
        }
    }
}
