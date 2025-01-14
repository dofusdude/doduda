namespace Ankama.Localization.Utils
{
    /// <summary>
    /// A Least Recently Used (LRU) cache implementation.
    /// </summary>
    public class LRUCache<TKey, TValue>
    {
        private readonly Dictionary<TKey, LinkedListNode<CacheItem>> _cache;
        private readonly LinkedList<CacheItem> _recentAccesses;
        private readonly int _maximumItemCount;
        private readonly Action<TKey, TValue> _onRemoveElement;

        /// <summary>
        /// Initializes a new instance of the <see cref="LRUCache{TKey, TValue}"/> class.
        /// </summary>
        /// <param name="maximumItemCount">The maximum number of items allowed in the cache.</param>
        /// <param name="onRemoveElement">An optional action to execute when an item is removed.</param>
        public LRUCache(int maximumItemCount = 512, Action<TKey, TValue> onRemoveElement = null)
        {
            _cache = new Dictionary<TKey, LinkedListNode<CacheItem>>(maximumItemCount);
            _recentAccesses = new LinkedList<CacheItem>();
            _maximumItemCount = maximumItemCount;
            _onRemoveElement = onRemoveElement;
        }

        /// <summary>
        /// Tries to retrieve the value associated with the specified key.
        /// </summary>
        /// <param name="key">The key of the value to retrieve.</param>
        /// <param name="value">The value associated with the key, if found.</param>
        /// <returns><c>true</c> if the key exists in the cache; otherwise, <c>false</c>.</returns>
        public bool TryGetValue(TKey key, out TValue value)
        {
            if (_cache.TryGetValue(key, out var node))
            {
                // Move the accessed item to the front of the linked list
                _recentAccesses.Remove(node);
                _recentAccesses.AddFirst(node);
                value = node.Value.Value;
                return true;
            }

            value = default;
            return false;
        }

        /// <summary>
        /// Tries to add a key-value pair to the cache.
        /// </summary>
        /// <param name="key">The key of the item to add.</param>
        /// <param name="value">The value of the item to add.</param>
        /// <returns><c>true</c> if the item was added; <c>false</c> if the key already exists.</returns>
        public bool TryAdd(TKey key, TValue value)
        {
            if (_cache.ContainsKey(key))
            {
                return false;
            }

            // If the cache is full, remove the least recently used item
            if (_cache.Count >= _maximumItemCount)
            {
                RemoveLeastRecentlyUsedItem();
            }

            var cacheItem = new CacheItem(key, value);
            var node = new LinkedListNode<CacheItem>(cacheItem);

            _cache[key] = node;
            _recentAccesses.AddFirst(node);

            return true;
        }

        /// <summary>
        /// Removes the least recently used item from the cache.
        /// </summary>
        private void RemoveLeastRecentlyUsedItem()
        {
            var lastNode = _recentAccesses.Last;
            if (lastNode != null)
            {
                _recentAccesses.RemoveLast();
                _cache.Remove(lastNode.Value.Key);
                _onRemoveElement?.Invoke(lastNode.Value.Key, lastNode.Value.Value);
            }
        }

        /// <summary>
        /// Represents an item in the cache.
        /// </summary>
        private class CacheItem
        {
            public TKey Key { get; }
            public TValue Value { get; }

            public CacheItem(TKey key, TValue value)
            {
                Key = key;
                Value = value;
            }
        }
    }
}
