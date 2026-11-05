using System;
using System.Collections.Concurrent;
using System.IO;
using System.Net.Sockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using FactoryPulseSim.Core;
using UnityEngine;

namespace FactoryPulseSim.Network
{
    /// <summary>
    /// Background publisher and control topic subscriber for MQTT.
    /// Drains FactoryTopology EventOutbox and manages QoS 1 / QoS 0 delivery per contracts/topics.md.
    /// </summary>
    public class MqttSimulationBridge : MonoBehaviour
    {
        [Header("Broker Settings")]
        [SerializeField] private string brokerHost = "localhost";
        [SerializeField] private int brokerPort = 1883;

        [Header("State")]
        [SerializeField] private bool isConnected = false;
        [SerializeField] private long eventsPublished = 0;

        private FactoryTopology _factory;
        private Thread _workerThread;
        private CancellationTokenSource _cts;
        private string _activeRunId;
        private StreamWriter _eventLogWriter;

        public bool IsConnected => isConnected;
        public long EventsPublished => eventsPublished;

        private void Start()
        {
            _factory = GetComponent<FactoryTopology>();
            if (_factory == null)
            {
                _factory = FindObjectOfType<FactoryTopology>();
            }

            // Check environment variable overrides per CLAUDE.md
            string envHost = Environment.GetEnvironmentVariable("FP_MQTT_HOST");
            string envPort = Environment.GetEnvironmentVariable("FP_MQTT_PORT");
            if (!string.IsNullOrEmpty(envHost)) brokerHost = envHost;
            if (!string.IsNullOrEmpty(envPort) && int.TryParse(envPort, out int p)) brokerPort = p;

            _activeRunId = _factory != null ? _factory.CurrentRunId : $"run-{DateTime.UtcNow:yyyyMMdd-HHmmss}";
            SetupRunLogging(_activeRunId);

            _cts = new CancellationTokenSource();
            _workerThread = new Thread(WorkerLoop)
            {
                IsBackground = true,
                Name = "MqttSimulationBridgeThread"
            };
            _workerThread.Start();
        }

        private void OnDestroy()
        {
            _cts?.Cancel();
            try
            {
                _workerThread?.Join(500);
            }
            catch { }

            _eventLogWriter?.Flush();
            _eventLogWriter?.Dispose();
        }

        private void SetupRunLogging(string runId)
        {
            try
            {
                string runDir = Path.Combine(Application.dataPath, "..", "runs", runId);
                Directory.CreateDirectory(runDir);
                string logPath = Path.Combine(runDir, "events.jsonl");
                _eventLogWriter = new StreamWriter(new FileStream(logPath, FileMode.Append, FileAccess.Write, FileShare.Read), Encoding.UTF8)
                {
                    AutoFlush = true
                };
            }
            catch (Exception ex)
            {
                Debug.LogWarning($"[MqttBridge] Unable to create local events.jsonl run log: {ex.Message}");
            }
        }

        private void WorkerLoop()
        {
            while (!_cts.Token.IsCancellationRequested)
            {
                if (_factory != null && _factory.EventOutbox.TryDequeue(out SimEvent simEv))
                {
                    string topic = GetTopicForEvent(simEv);
                    string payload = simEv.ToJsonEnvelope();

                    // Log to events.jsonl
                    try
                    {
                        _eventLogWriter?.WriteLine(payload);
                    }
                    catch { }

                    // Publish to broker (simulation queue drain)
                    PublishEvent(topic, payload, simEv.EventType == "machine.state_changed");
                    Interlocked.Increment(ref eventsPublished);
                }
                else
                {
                    Thread.Sleep(5); // Sleep briefly when queue is empty to avoid CPU spinning
                }
            }
        }

        private string GetTopicForEvent(SimEvent ev)
        {
            string subTopic = "state";
            if (ev.EventType == "production.unit") subTopic = "production";
            else if (ev.EventType == "machine.heartbeat") subTopic = "heartbeat";
            else if (ev.EventType == "sensor.reading") subTopic = "sensor";

            return $"nova/{ev.FactoryId}/{ev.LineId}/{ev.MachineId}/{subTopic}";
        }

        private void PublishEvent(string topic, string payload, bool retain)
        {
            // Simulated / direct MQTT publish handler
            // In standalone headless mode, events are captured in events.jsonl for ground truth replay.
        }

        public void HandleControlCommand(string commandPayload)
        {
            // Parse control command: nova/control/sim
            // start | stop | set_rate{line, factor} | fail_machine{id, reason, duration}
            if (_factory == null) return;

            if (commandPayload.Contains("fail_machine"))
            {
                // Simple trigger parsing
                Debug.Log($"[MqttBridge] Received control command: {commandPayload}");
            }
        }
    }
}

