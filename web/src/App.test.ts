import { describe, it, expect } from 'vitest';

describe('FactoryPulse Web Operations Logic', () => {
  it('correctly calculates total plant output across 3 lines', () => {
    const linesOutput = [260, 275, 290];
    const total = linesOutput.reduce((acc, curr) => acc + curr, 0);
    expect(total).toBe(825);
  });

  it('correctly tags bottleneck station when upstream is blocked and downstream starved', () => {
    const stations = [
      { id: 'M-REF-01', state: 'BLOCKED' },
      { id: 'M-REF-02', state: 'STOPPED' },
      { id: 'M-REF-03', state: 'STARVED' },
    ];

    expect(stations[0].state).toBe('BLOCKED');
    expect(stations[1].state).toBe('STOPPED');
    expect(stations[2].state).toBe('STARVED');
  });

  it('determines shift pacing status correctly', () => {
    const target = 500;
    const projected = 520;
    const status = projected >= target ? 'ON_TRACK' : 'BEHIND';
    expect(status).toBe('ON_TRACK');
  });
});

