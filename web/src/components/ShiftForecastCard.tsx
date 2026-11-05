import React from 'react';
import type { ShiftForecastData } from '../types';

interface ShiftForecastCardProps {
  forecast: ShiftForecastData;
}

export const ShiftForecastCard: React.FC<ShiftForecastCardProps> = ({ forecast }) => {
  const getStatusColor = () => {
    switch (forecast.status) {
      case 'ON_TRACK':
        return { bg: '#14532d', text: '#4ade80', label: 'ON TRACK' };
      case 'AT_RISK':
        return { bg: '#78350f', text: '#fde047', label: 'AT RISK' };
      case 'BEHIND':
        return { bg: '#7f1d1d', text: '#fca5a5', label: 'BEHIND SCHEDULE' };
      case 'INSUFFICIENT_DATA':
      default:
        return { bg: '#334155', text: '#94a3b8', label: 'SHIFT STARTING (WARMING UP)' };
    }
  };

  const statusInfo = getStatusColor();

  return (
    <div style={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', padding: '20px', marginBottom: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <div>
          <h2 style={{ margin: 0, fontSize: '18px', color: '#f8fafc' }}>Shift Production Forecast</h2>
          <span style={{ fontSize: '12px', color: '#94a3b8' }}>{forecast.assumptionWindow}</span>
        </div>
        <span style={{
          padding: '4px 12px',
          borderRadius: '9999px',
          fontWeight: 'bold',
          fontSize: '12px',
          backgroundColor: statusInfo.bg,
          color: statusInfo.text
        }}>
          {statusInfo.label}
        </span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '16px' }}>
        <div>
          <div style={{ fontSize: '12px', color: '#94a3b8' }}>Target Shift Volume</div>
          <div style={{ fontSize: '22px', fontWeight: 'bold', color: '#f8fafc' }}>{forecast.shiftTarget} units</div>
        </div>
        <div>
          <div style={{ fontSize: '12px', color: '#94a3b8' }}>Projected Shift Total</div>
          <div style={{ fontSize: '22px', fontWeight: 'bold', color: '#38bdf8' }}>{forecast.projectedTotal} units</div>
        </div>
        <div>
          <div style={{ fontSize: '12px', color: '#94a3b8' }}>Current Pace</div>
          <div style={{ fontSize: '22px', fontWeight: 'bold', color: '#f8fafc' }}>{forecast.currentHourlyRate.toFixed(1)} u/hr</div>
        </div>
        <div>
          <div style={{ fontSize: '12px', color: '#94a3b8' }}>Required Pace</div>
          <div style={{ fontSize: '22px', fontWeight: 'bold', color: '#f59e0b' }}>{forecast.requiredHourlyRate.toFixed(1)} u/hr</div>
        </div>
        <div>
          <div style={{ fontSize: '12px', color: '#94a3b8' }}>Shift Remaining</div>
          <div style={{ fontSize: '22px', fontWeight: 'bold', color: '#cbd5e1' }}>{forecast.hoursLeft.toFixed(1)} hrs</div>
        </div>
      </div>
    </div>
  );
};

