import { useState } from 'react';
import { Header } from './components/Header';
import { Overview } from './components/Overview';
import { LineView } from './components/LineView';
import { IncidentBoard } from './components/IncidentBoard';
import { ShiftForecastCard } from './components/ShiftForecastCard';
import { useLiveSocket } from './hooks/useLiveSocket';
import type { LineData, IncidentData, ShiftForecastData } from './types';

export function App() {
  const { status } = useLiveSocket(() => {
    // live update callback
  });

  const [incidents, setIncidents] = useState<IncidentData[]>([
    {
      id: 'inc-01',
      machineId: 'M-REF-02',
      status: 'open',
      openedAt: new Date(Date.now() - 1000 * 60 * 8).toISOString(),
      rootCause: 'HEATER_FAILURE',
      sourceEventId: 'ev-9988'
    }
  ]);

  const [forecast] = useState<ShiftForecastData>({
    shiftTarget: 500,
    unitsProduced: 260,
    unitsRemaining: 240,
    hoursLeft: 4.0,
    currentHourlyRate: 65.0,
    requiredHourlyRate: 60.0,
    projectedTotal: 520,
    status: 'ON_TRACK',
    assumptionWindow: 'Rolling 30-60 min production rate'
  });

  const [lines] = useState<LineData[]>([
    {
      lineId: 'line-refrig-01',
      name: 'Line 1: Refrigerator Assembly',
      productType: 'refrigerator-standard',
      targetUnits: 500,
      actualUnits: 260,
      oee: 0.84,
      bottleneckMachineId: 'M-REF-02',
      stations: [
        {
          machineId: 'M-REF-01',
          name: 'Sheet Press',
          stationOrder: 1,
          state: 'BLOCKED',
          upstreamCauseId: 'M-REF-02',
          idealCycleMs: 50000,
          unitsProduced: 285,
          oee: { availability: 0.88, performance: 0.95, quality: 0.99, overall: 0.83 }
        },
        {
          machineId: 'M-REF-02',
          name: 'Foaming Station',
          stationOrder: 2,
          state: 'STOPPED',
          reasonCode: 'HEATER_FAILURE',
          idealCycleMs: 52000,
          unitsProduced: 265,
          oee: { availability: 0.78, performance: 0.92, quality: 0.98, overall: 0.70 }
        },
        {
          machineId: 'M-REF-03',
          name: 'Compressor Mount',
          stationOrder: 3,
          state: 'STARVED',
          upstreamCauseId: 'M-REF-02',
          idealCycleMs: 53000,
          unitsProduced: 262,
          oee: { availability: 0.82, performance: 0.94, quality: 0.99, overall: 0.76 }
        },
        {
          machineId: 'M-REF-04',
          name: 'Leak & QC Test',
          stationOrder: 4,
          state: 'RUNNING',
          idealCycleMs: 54000,
          unitsProduced: 260,
          oee: { availability: 0.89, performance: 0.95, quality: 0.98, overall: 0.83 }
        }
      ],
      buffers: [
        { bufferId: 'B-REF-01', capacity: 10, currentCount: 10 },
        { bufferId: 'B-REF-02', capacity: 8, currentCount: 0 },
        { bufferId: 'B-REF-03', capacity: 12, currentCount: 2 }
      ]
    },
    {
      lineId: 'line-wash-01',
      name: 'Line 2: Washing Machine Assembly',
      productType: 'washer-frontload',
      targetUnits: 500,
      actualUnits: 275,
      oee: 0.88,
      stations: [
        {
          machineId: 'M-WASH-01',
          name: 'Drum Stamping',
          stationOrder: 1,
          state: 'RUNNING',
          idealCycleMs: 48000,
          unitsProduced: 290,
          oee: { availability: 0.94, performance: 0.96, quality: 0.99, overall: 0.89 }
        },
        {
          machineId: 'M-WASH-02',
          name: 'Tub Subassembly',
          stationOrder: 2,
          state: 'RUNNING',
          idealCycleMs: 51000,
          unitsProduced: 282,
          oee: { availability: 0.92, performance: 0.95, quality: 0.99, overall: 0.86 }
        },
        {
          machineId: 'M-WASH-03',
          name: 'Direct Drive Mount',
          stationOrder: 3,
          state: 'RUNNING',
          idealCycleMs: 53000,
          unitsProduced: 278,
          oee: { availability: 0.91, performance: 0.94, quality: 0.99, overall: 0.85 }
        },
        {
          machineId: 'M-WASH-04',
          name: 'Spin & Balance Test',
          stationOrder: 4,
          state: 'RUNNING',
          idealCycleMs: 54000,
          unitsProduced: 275,
          oee: { availability: 0.93, performance: 0.95, quality: 0.98, overall: 0.87 }
        }
      ],
      buffers: [
        { bufferId: 'B-WASH-01', capacity: 12, currentCount: 4 },
        { bufferId: 'B-WASH-02', capacity: 8, currentCount: 3 },
        { bufferId: 'B-WASH-03', capacity: 10, currentCount: 5 }
      ]
    },
    {
      lineId: 'line-micro-01',
      name: 'Line 3: Microwave Assembly',
      productType: 'microwave-countertop',
      targetUnits: 500,
      actualUnits: 290,
      oee: 0.91,
      stations: [
        {
          machineId: 'M-MIC-01',
          name: 'Chassis Press',
          stationOrder: 1,
          state: 'RUNNING',
          idealCycleMs: 46000,
          unitsProduced: 305,
          oee: { availability: 0.95, performance: 0.96, quality: 0.99, overall: 0.90 }
        },
        {
          machineId: 'M-MIC-02',
          name: 'Cavity Welder',
          stationOrder: 2,
          state: 'RUNNING',
          idealCycleMs: 50000,
          unitsProduced: 298,
          oee: { availability: 0.94, performance: 0.95, quality: 0.99, overall: 0.88 }
        },
        {
          machineId: 'M-MIC-03',
          name: 'Magnetron Mount',
          stationOrder: 3,
          state: 'RUNNING',
          idealCycleMs: 52000,
          unitsProduced: 294,
          oee: { availability: 0.93, performance: 0.95, quality: 0.99, overall: 0.87 }
        },
        {
          machineId: 'M-MIC-04',
          name: 'Safety & Leak Test',
          stationOrder: 4,
          state: 'RUNNING',
          idealCycleMs: 54000,
          unitsProduced: 290,
          oee: { availability: 0.95, performance: 0.96, quality: 0.99, overall: 0.90 }
        }
      ],
      buffers: [
        { bufferId: 'B-MIC-01', capacity: 15, currentCount: 6 },
        { bufferId: 'B-MIC-02', capacity: 10, currentCount: 4 },
        { bufferId: 'B-MIC-03', capacity: 10, currentCount: 5 }
      ]
    }
  ]);

  const handleAcknowledge = (id: string) => {
    setIncidents((prev) =>
      prev.map((inc) =>
        inc.id === id
          ? {
              ...inc,
              status: 'acknowledged' as const,
              acknowledgedAt: new Date().toISOString(),
              assignee: 'Technician (You)'
            }
          : inc
      )
    );
  };

  const handleResolve = (id: string) => {
    setIncidents((prev) =>
      prev.map((inc) =>
        inc.id === id
          ? {
              ...inc,
              status: 'resolved' as const,
              resolvedAt: new Date().toISOString()
            }
          : inc
      )
    );
  };

  return (
    <div style={{ minHeight: '100vh', backgroundColor: '#0f172a', paddingBottom: '40px' }}>
      <Header status={status} activeShift="Shift 1 (06:00 - 14:00 UTC)" />

      <main style={{ maxWidth: '1400px', margin: '0 auto', padding: '24px' }}>
        <Overview
          lines={lines}
          incidents={incidents}
          totalProduced={lines.reduce((acc, l) => acc + l.actualUnits, 0)}
          totalTarget={1500}
        />

        <ShiftForecastCard forecast={forecast} />

        <IncidentBoard
          incidents={incidents}
          onAcknowledge={handleAcknowledge}
          onResolve={handleResolve}
        />

        <div style={{ marginTop: '32px' }}>
          <h2 style={{ fontSize: '18px', color: '#f8fafc', marginBottom: '16px' }}>
            Production Line Topology & Station Status
          </h2>
          {lines.map((line) => (
            <LineView key={line.lineId} line={line} />
          ))}
        </div>
      </main>
    </div>
  );
}

export default App;

