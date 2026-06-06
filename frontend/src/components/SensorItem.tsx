import React from 'react';
import { useDraggable } from '@dnd-kit/core';
import { SensorType } from '../types';
import { CSS } from '@dnd-kit/utilities';

interface SensorItemProps {
  sensor: SensorType;
}

export const SensorItem: React.FC<SensorItemProps> = ({ sensor }) => {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: `sensor-${sensor.id}`,
    data: { sensor, type: 'sensor' },
  });

  const style = {
    transform: CSS.Translate.toString(transform),
    opacity: isDragging ? 0.5 : 1,
    backgroundColor: sensor.color,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      className="p-3 rounded-lg text-white cursor-grab active:cursor-grabbing shadow-md hover:shadow-lg transition-shadow select-none"
    >
      <div className="font-semibold text-sm">{sensor.name}</div>
      <div className="text-xs opacity-90 mt-1">{sensor.dataType} · {sensor.regCount} 寄存器</div>
    </div>
  );
};
