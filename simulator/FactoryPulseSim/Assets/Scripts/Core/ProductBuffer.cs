using System;
using UnityEngine;

namespace FactoryPulseSim.Core
{
    /// <summary>
    /// Represents an inter-station physical buffer / conveyor queue between two machines.
    /// Manages capacity, push/pop operations, and provides starvation/blocking signals.
    /// </summary>
    public class ProductBuffer : MonoBehaviour
    {
        [SerializeField] private string bufferId = "B-01";
        [SerializeField] private int capacity = 10;
        [SerializeField] private int currentCount = 0;

        public event Action<int, int> OnCountChanged; // (currentCount, capacity)

        public string BufferId => bufferId;
        public int Capacity => capacity;
        public int CurrentCount => currentCount;
        public bool IsFull => currentCount >= capacity;
        public bool IsEmpty => currentCount <= 0;
        public float OccupancyRatio => capacity > 0 ? (float)currentCount / capacity : 0f;

        public void Initialize(string id, int cap, int initialCount = 0)
        {
            bufferId = id;
            capacity = Mathf.Max(1, cap);
            currentCount = Mathf.Clamp(initialCount, 0, capacity);
            OnCountChanged?.Invoke(currentCount, capacity);
        }

        /// <summary>
        /// Attempts to push a completed unit into the buffer.
        /// Returns false if the buffer is full (machine becomes BLOCKED).
        /// </summary>
        public bool TryPush()
        {
            if (currentCount >= capacity)
            {
                return false;
            }

            currentCount++;
            OnCountChanged?.Invoke(currentCount, capacity);
            return true;
        }

        /// <summary>
        /// Attempts to pop an input unit from the buffer.
        /// Returns false if the buffer is empty (machine becomes STARVED).
        /// </summary>
        public bool TryPop()
        {
            if (currentCount <= 0)
            {
                return false;
            }

            currentCount--;
            OnCountChanged?.Invoke(currentCount, capacity);
            return true;
        }

        public void ResetBuffer(int initialCount = 0)
        {
            currentCount = Mathf.Clamp(initialCount, 0, capacity);
            OnCountChanged?.Invoke(currentCount, capacity);
        }
    }
}
