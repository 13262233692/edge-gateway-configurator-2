export interface SensorType {
  id: number;
  name: string;
  description: string;
  dataType: string;
  regCount: number;
  color: string;
}

export interface GatewayTemplate {
  id: number;
  name: string;
  model: string;
  description: string;
}

export interface RegisterBinding {
  id: number;
  gatewayTemplateId: number;
  sensorTypeId: number;
  startAddress: number;
  label: string;
  sensorType: SensorType;
}

export interface BindingDraft {
  sensorType: SensorType;
  startAddress: number;
  label?: string;
}
