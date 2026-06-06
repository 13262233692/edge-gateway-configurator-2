import React, { useState, useEffect } from 'react';
import {
  DndContext,
  DragEndEvent,
  DragOverEvent,
  PointerSensor,
  useSensor,
  useSensors,
  DragStartEvent,
} from '@dnd-kit/core';
import { SensorType, GatewayTemplate, RegisterBinding, BindingDraft } from './types';
import { sensorApi, templateApi, bindingApi } from './api';
import { SensorItem } from './components/SensorItem';
import { RegisterCell } from './components/RegisterCell';

const START_ADDR = 40001;
const END_ADDR = 40100;
const TOTAL_CELLS = END_ADDR - START_ADDR + 1;

function App() {
  const [sensors, setSensors] = useState<SensorType[]>([]);
  const [templates, setTemplates] = useState<GatewayTemplate[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<number | null>(null);
  const [bindings, setBindings] = useState<RegisterBinding[]>([]);
  const [activeSensor, setActiveSensor] = useState<SensorType | null>(null);
  const [overAddress, setOverAddress] = useState<number | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const sensorsList = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  );

  useEffect(() => {
    loadData();
  }, []);

  useEffect(() => {
    if (selectedTemplate) {
      loadBindings(selectedTemplate);
    }
  }, [selectedTemplate]);

  const loadData = async () => {
    try {
      const [sensorsRes, templatesRes] = await Promise.all([
        sensorApi.getAll(),
        templateApi.getAll(),
      ]);
      setSensors(sensorsRes.data);
      setTemplates(templatesRes.data);
      if (templatesRes.data.length > 0) {
        setSelectedTemplate(templatesRes.data[0].id);
      }
    } catch (err) {
      showMessage('error', '加载数据失败');
    }
  };

  const loadBindings = async (templateId: number) => {
    try {
      const res = await bindingApi.getAll(templateId);
      setBindings(res.data);
    } catch (err) {
      showMessage('error', '加载绑定数据失败');
    }
  };

  const showMessage = (type: 'success' | 'error', text: string) => {
    setMessage({ type, text });
    setTimeout(() => setMessage(null), 3000);
  };

  const getOccupiedMap = () => {
    const map = new Map<number, { sensor: SensorType; bindingId: number }>();
    bindings.forEach((b) => {
      const regCount = b.sensorType.regCount > 0 ? b.sensorType.regCount : 1;
      for (let i = 0; i < regCount; i++) {
        map.set(b.startAddress + i, { sensor: b.sensorType, bindingId: b.id });
      }
    });
    return map;
  };

  const checkValidDrop = (address: number, sensor: SensorType): { valid: boolean; reason?: string } => {
    if (sensor.regCount <= 0) {
      return { valid: false, reason: '传感器寄存器数量配置无效' };
    }
    const endAddr = address + sensor.regCount - 1;
    if (endAddr > END_ADDR) {
      return { valid: false, reason: `地址范围 ${address}-${endAddr} 超出 40100 边界` };
    }

    const occupied = getOccupiedMap();
    for (let i = 0; i < sensor.regCount; i++) {
      const addr = address + i;
      if (occupied.has(addr)) {
        const conflict = occupied.get(addr)!;
        const conflictEnd = conflict.sensor.regCount > 0
          ? conflict.sensor.regCount
          : 1;
        return {
          valid: false,
          reason: `地址 ${addr} 与传感器 \"${conflict.sensor.name}\" 重叠`,
        };
      }
    }
    return { valid: true };
  };

  const handleDragStart = (event: DragStartEvent) => {
    const data = event.active.data.current;
    if (data?.type === 'sensor') {
      setActiveSensor(data.sensor);
    }
  };

  const handleDragOver = (event: DragOverEvent) => {
    const overId = event.over?.id as string;
    if (overId?.startsWith('cell-')) {
      const addr = parseInt(overId.replace('cell-', ''));
      setOverAddress(addr);
    } else {
      setOverAddress(null);
    }
  };

  const handleDragEnd = (event: DragEndEvent) => {
    setActiveSensor(null);
    setOverAddress(null);

    const { active, over } = event;
    const activeData = active.data.current;

    if (!over || activeData?.type !== 'sensor') return;

    const overId = over.id as string;
    if (!overId.startsWith('cell-')) return;

    const address = parseInt(overId.replace('cell-', ''));
    const sensor = activeData.sensor as SensorType;

    const { valid, reason } = checkValidDrop(address, sensor);
    if (!valid) {
      showMessage('error', reason || '地址无效或与已有传感器重叠');
      return;
    }

    if (!selectedTemplate) {
      showMessage('error', '请先选择网关模板');
      return;
    }

    bindingApi
      .create({
        gatewayTemplateId: selectedTemplate,
        sensorTypeId: sensor.id,
        startAddress: address,
        label: sensor.name,
      })
      .then(() => {
        loadBindings(selectedTemplate);
        showMessage('success', `已绑定 ${sensor.name} 到地址 ${address}`);
      })
      .catch((err) => {
        const msg = err.response?.data?.error || '绑定失败';
        showMessage('error', msg);
      });
  };

  const handleRemoveBinding = (bindingId: number) => {
    bindingApi
      .delete(bindingId)
      .then(() => {
        if (selectedTemplate) loadBindings(selectedTemplate);
        showMessage('success', '已移除传感器绑定');
      })
      .catch(() => showMessage('error', '移除失败'));
  };

  const handleClearAll = () => {
    if (!selectedTemplate) return;
    if (!window.confirm('确定要清空所有绑定吗？')) return;
    bindingApi
      .saveAll({ gatewayTemplateId: selectedTemplate, bindings: [] })
      .then(() => {
        setBindings([]);
        showMessage('success', '已清空所有绑定');
      })
      .catch(() => showMessage('error', '清空失败'));
  };

  const occupiedMap = getOccupiedMap();
  const addresses = Array.from({ length: TOTAL_CELLS }, (_, i) => START_ADDR + i);

  return (
    <div className="min-h-screen bg-slate-100">
      <header className="bg-slate-800 text-white px-6 py-4 shadow-lg">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <h1 className="text-xl font-bold">边缘网关配置下发系统</h1>
          <div className="flex items-center gap-4">
            <label className="text-sm">网关模板：</label>
            <select
              className="px-3 py-1.5 rounded bg-slate-700 text-white text-sm border border-slate-600"
              value={selectedTemplate || ''}
              onChange={(e) => setSelectedTemplate(parseInt(e.target.value))}
            >
              {templates.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </select>
          </div>
        </div>
      </header>

      {message && (
        <div
          className={`fixed top-20 right-6 z-50 px-4 py-2 rounded shadow-lg text-white ${
            message.type === 'success' ? 'bg-green-500' : 'bg-red-500'
          }`}
        >
          {message.text}
        </div>
      )}

      <div className="max-w-7xl mx-auto p-6">
        <div className="flex gap-6">
          <div className="w-72 flex-shrink-0">
            <div className="bg-white rounded-lg shadow p-4">
              <h2 className="font-semibold text-gray-700 mb-4">传感器类型库</h2>
              <div className="space-y-3">
                {sensors.map((s) => (
                  <SensorItem key={s.id} sensor={s} />
                ))}
              </div>
            </div>

            <div className="bg-white rounded-lg shadow p-4 mt-4">
              <h3 className="font-semibold text-gray-700 mb-3">操作</h3>
              <button
                onClick={handleClearAll}
                className="w-full py-2 px-4 bg-red-500 text-white rounded hover:bg-red-600 transition text-sm"
              >
                清空所有绑定
              </button>
              <div className="mt-4 text-xs text-gray-500 space-y-1">
                <p>· 寄存器地址范围: 40001 - 40100</p>
                <p>· 拖拽传感器到右侧寄存器格子</p>
                <p>· 系统自动检测地址重叠</p>
              </div>
            </div>
          </div>

          <div className="flex-1">
            <DndContext
              sensors={sensorsList}
              onDragStart={handleDragStart}
              onDragOver={handleDragOver}
              onDragEnd={handleDragEnd}
            >
              <div className="bg-white rounded-lg shadow p-4">
                <h2 className="font-semibold text-gray-700 mb-4">
                  保持寄存器映射视图 (40001 - 40100)
                </h2>
                <div className="grid grid-cols-10 gap-1.5">
                  {addresses.map((addr) => {
                    const occupied = occupiedMap.get(addr);
                    const isOver =
                      activeSensor &&
                      overAddress !== null &&
                      addr >= overAddress &&
                      addr < overAddress + activeSensor.regCount;
                    const dropResult = activeSensor
                      ? checkValidDrop(overAddress || addr, activeSensor)
                      : { valid: false };
                    return (
                      <RegisterCell
                        key={addr}
                        address={addr}
                        occupied={!!occupied}
                        sensor={occupied?.sensor}
                        bindingId={occupied?.bindingId}
                        isOver={isOver}
                        isValidDrop={dropResult.valid}
                        onRemove={handleRemoveBinding}
                      />
                    );
                  })}
                </div>
              </div>
            </DndContext>

            <div className="bg-white rounded-lg shadow p-4 mt-4">
              <h3 className="font-semibold text-gray-700 mb-3">已绑定传感器列表</h3>
              {bindings.length === 0 ? (
                <p className="text-gray-400 text-sm">暂无绑定</p>
              ) : (
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {bindings.map((b) => (
                    <div
                      key={b.id}
                      className="flex items-center justify-between p-2 rounded border"
                      style={{
                        borderColor: b.sensorType.color,
                        backgroundColor: b.sensorType.color + '15',
                      }}
                    >
                      <div>
                        <span className="font-medium text-sm">{b.sensorType.name}</span>
                        <span className="text-gray-500 text-xs ml-2">
                          地址: {b.startAddress} - {b.startAddress + b.sensorType.regCount - 1}
                        </span>
                        <span className="text-gray-400 text-xs ml-2">
                          ({b.sensorType.regCount} 寄存器 · {b.sensorType.dataType})
                        </span>
                      </div>
                      <button
                        onClick={() => handleRemoveBinding(b.id)}
                        className="text-red-500 hover:text-red-700 text-sm px-2"
                      >
                        删除
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
