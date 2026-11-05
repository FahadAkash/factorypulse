using System;
using UnityEngine;

namespace FactoryPulseSim.Core
{
    /// <summary>
    /// Deterministic failure injector implementing MTBF/MTTR exponential distributions
    /// using seeded System.Random per machine. Supports manual fault injection.
    /// </summary>
    public class FailureInjector
    {
        private readonly System.Random _rng;
        private readonly float _mtbfSeconds;
        private readonly float _mttrSeconds;
        private readonly string[] _reasonCodes;

        private double _nextFailureSimTime = 0.0;
        private double _faultExpirySimTime = 0.0;
        private bool _isFaultActive = false;
        private string _activeReasonCode = null;

        public bool IsFaultActive => _isFaultActive;
        public string ActiveReasonCode => _activeReasonCode;
        public double NextFailureSimTime => _nextFailureSimTime;

        public FailureInjector(int masterSeed, string machineId, float mtbf, float mttr, string[] reasonCodes)
        {
            // Seed per machine = Hash(masterSeed, machineId)
            int machineSeed = unchecked(masterSeed * 397 ^ (machineId != null ? machineId.GetHashCode() : 0));
            _rng = new System.Random(machineSeed);

            _mtbfSeconds = Mathf.Max(10f, mtbf);
            _mttrSeconds = Mathf.Max(5f, mttr);
            _reasonCodes = (reasonCodes != null && reasonCodes.Length > 0)
                ? reasonCodes
                : new string[] { "MOTOR_OVERLOAD", "JAM", "SENSOR_FAULT" };

            // Schedule first failure using exponential distribution
            _nextFailureSimTime = SampleExponential(_mtbfSeconds);
        }

        /// <summary>
        /// Evaluates failure status for the current simulation time.
        /// Returns true if machine is in a STOPPED failure state.
        /// </summary>
        public bool UpdateFailureStatus(double currentSimSeconds, out string reasonCode)
        {
            if (_isFaultActive)
            {
                if (currentSimSeconds >= _faultExpirySimTime)
                {
                    // Fault repaired
                    _isFaultActive = false;
                    _activeReasonCode = null;
                    // Schedule next failure
                    _nextFailureSimTime = currentSimSeconds + SampleExponential(_mtbfSeconds);
                    reasonCode = null;
                    return false;
                }

                reasonCode = _activeReasonCode;
                return true;
            }

            // Check if it is time for an unplanned failure
            if (currentSimSeconds >= _nextFailureSimTime)
            {
                _isFaultActive = true;
                _activeReasonCode = _reasonCodes[_rng.Next(_reasonCodes.Length)];
                float repairDuration = SampleExponential(_mttrSeconds);
                _faultExpirySimTime = currentSimSeconds + repairDuration;

                reasonCode = _activeReasonCode;
                return true;
            }

            reasonCode = null;
            return false;
        }

        /// <summary>
        /// Injects a targeted fault directly (e.g. from control topic or scenario).
        /// </summary>
        public void InjectManualFault(double currentSimSeconds, string reason, float durationSeconds)
        {
            _isFaultActive = true;
            _activeReasonCode = !string.IsNullOrEmpty(reason) ? reason : "MANUAL_FAULT_INJECTED";
            _faultExpirySimTime = currentSimSeconds + Mathf.Max(1f, durationSeconds);
        }

        public void ClearFault(double currentSimSeconds)
        {
            _isFaultActive = false;
            _activeReasonCode = null;
            _nextFailureSimTime = currentSimSeconds + SampleExponential(_mtbfSeconds);
        }

        private float SampleExponential(float mean)
        {
            // Inverse transform sampling: -mean * ln(1 - U)
            double u = _rng.NextDouble();
            if (u >= 0.999999) u = 0.99999;
            return (float)(-mean * Math.Log(1.0 - u));
        }
    }
}

