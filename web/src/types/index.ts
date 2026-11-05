export type MachineState =
  | 'RUNNING'
  | 'STOPPED'
  | 'STARVED'
  | 'BLOCKED'
  | 'IDLE'
  | 'MAINTENANCE';

export interface StationData {
  machineId: string;
  name: string;
  stationOrder: number;
  state: MachineState;
  reasonCode?: string;
  upstreamCauseId?: string;
  idealCycleMs: number;
  unitsProduced: number;
  oee: {
    availability: number;
    performance: number;
    quality: number;
    overall: number;
  };
}

export interface BufferData {
  bufferId: string;
  capacity: number;
  currentCount: number;
}

export interface LineData {
  lineId: string;
  name: string;
  productType: string;
  targetUnits: number;
  actualUnits: number;
  oee: number;
  stations: StationData[];
  buffers: BufferData[];
  bottleneckMachineId?: string;
}

export interface IncidentData {
  id: string;
  machineId: string;
  status: 'open' | 'acknowledged' | 'resolved';
  openedAt: string;
  acknowledgedAt?: string;
  resolvedAt?: string;
  assignee?: string;
  rootCause?: string;
  sourceEventId: string;
}

export interface ShiftForecastData {
  shiftTarget: number;
  unitsProduced: number;
  unitsRemaining: number;
  hoursLeft: number;
  currentHourlyRate: number;
  requiredHourlyRate: number;
  projectedTotal: number;
  status: 'ON_TRACK' | 'AT_RISK' | 'BEHIND' | 'INSUFFICIENT_DATA';
  assumptionWindow: string;
}

export type ConnectionStatus = 'LIVE' | 'RECONNECTING' | 'STALE';

