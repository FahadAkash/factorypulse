using System;
using UnityEngine;

namespace FactoryPulseSim.Core
{
    /// <summary>
    /// Deterministic simulation clock advancing in discrete fixed-time steps.
    /// Maps simulation seconds to UTC DateTime starting from an anchor epoch.
    /// </summary>
    public class SimClock : MonoBehaviour
    {
        [Tooltip("Simulation fixed step in seconds (e.g. 0.1s = 100ms)")]
        [SerializeField] private float stepDeltaSeconds = 0.1f;

        [Tooltip("Simulation acceleration factor (1x to 60x)")]
        [Range(1f, 60f)]
        [SerializeField] private float timeScale = 1.0f;

        private double _elapsedSimSeconds = 0.0;
        private DateTime _epochStartTime;
        private bool _isRunning = false;

        public double ElapsedSimSeconds => _elapsedSimSeconds;
        public float StepDeltaSeconds => stepDeltaSeconds;
        public float TimeScale => timeScale;
        public bool IsRunning => _isRunning;

        public DateTime CurrentSimTime => _epochStartTime.AddSeconds(_elapsedSimSeconds);

        public void Initialize(DateTime? anchorEpoch = null, float scale = 1.0f)
        {
            _epochStartTime = anchorEpoch ?? DateTime.UtcNow;
            _elapsedSimSeconds = 0.0;
            timeScale = scale;
            _isRunning = true;
        }

        public void SetTimeScale(float scale)
        {
            timeScale = Mathf.Clamp(scale, 1f, 60f);
        }

        public void Pause() => _isRunning = false;
        public void Resume() => _isRunning = true;

        private void FixedUpdate()
        {
            if (!_isRunning) return;

            // Advance simulation clock by scaled fixed delta
            _elapsedSimSeconds += stepDeltaSeconds * timeScale;
        }
    }
}
