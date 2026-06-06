import React from 'react';
import { useDroppable } from '@dnd-kit/core';
import { SensorType } from '../types';

interface RegisterCellProps {
  address: number;
  occupied: boolean;
  sensor?: SensorType;
  isOver: boolean;
  isValidDrop: boolean;
  bindingId?: number;
  onRemove?: (bindingId: number) => void;
}

export const RegisterCell: React.FC<RegisterCellProps> = ({
  address,
  occupied,
  sensor,
  isOver,
  isValidDrop,
  bindingId,
  onRemove,
}) => {
  const { setNodeRef } = useDroppable({
    id: `cell-${address}`,
    data: { address, type: 'cell' },
  });

  let bgClass = 'bg-white';
  if (isOver) {
    bgClass = isValidDrop ? 'bg-green-100 border-green-500' : 'bg-red-100 border-red-500';
  } else if (occupied) {
    bgClass = 'bg-gray-100';
  }

  return (
    <div
      ref={setNodeRef}
      className={`relative w-16 h-16 border border-gray-300 rounded flex flex-col items-center justify-center text-xs transition-colors ${bgClass}`}
      style={sensor ? { backgroundColor: sensor.color + '30', borderColor: sensor.color } : {}}
    >
      <span className="font-mono text-gray-600">{address}</span>
      {sensor && (
        <>
          <span className="text-xs font-medium mt-1 truncate px-1" style={{ color: sensor.color }}>
            {sensor.name.length > 6 ? sensor.name.slice(0, 6) + '…' : sensor.name}
          </span>
          {bindingId && onRemove && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onRemove(bindingId);
              }}
              className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white rounded-full text-xs flex items-center justify-center hover:bg-red-600"
            >
              ×
            </button>
          )}
        </>
      )}
    </div>
  );
};
