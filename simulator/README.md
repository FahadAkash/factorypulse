# FactoryPulse — Unity Factory Simulator Setup Guide

## Overview
The simulator models Nova Appliances Ltd (`factory-01`):
- **3 Production Lines**: Refrigerator (`line-refrig-01`), Washing Machine (`line-wash-01`), Microwave (`line-micro-01`).
- **12 Machines**: 4 machines per line connected in series via 3 intermediate `ProductBuffer` queues.
- **Assembly Definition**: `Assets/Scripts/FactoryPulseSim.asmdef`.

---

## Unity Editor Setup Checklist (M3)

### Step 1: Open Project in Unity
1. Open **Unity Hub**.
2. Add project pointing to `simulator/FactoryPulseSim`. Recommended version: **Unity 2022.3 LTS** or **Unity 6**.
3. Allow Unity to compile `FactoryPulseSim.asmdef`.

### Step 2: Create ScriptableObject Assets
In the Unity Project window under `Assets/Config/`:
1. **MachineConfigs**:
   - Right-click -> `Create` -> `FactoryPulse` -> `Machine Config`.
   - Create 12 assets according to `docs/business/discovery.md`:
     - Line 1: `M-REF-01` (Sheet Press), `M-REF-02` (Foaming), `M-REF-03` (Compressor), `M-REF-04` (Leak & QC).
     - Line 2: `M-WASH-01` (Drum Stamping), `M-WASH-02` (Tub Subassembly), `M-WASH-03` (Drive Mount), `M-WASH-04` (Spin Test).
     - Line 3: `M-MIC-01` (Chassis Press), `M-MIC-02` (Cavity Welder), `M-MIC-03` (Magnetron Mount), `M-MIC-04` (Safety Test).
2. **LineConfigs**:
   - Right-click -> `Create` -> `FactoryPulse` -> `Line Config`.
   - Create 3 assets: `Line_Refrigerator`, `Line_Washer`, `Line_Microwave`.
   - Assign the corresponding 4 `MachineConfig` assets and buffer capacities (e.g. `[10, 8, 12]`).
3. **FactoryConfig**:
   - Right-click -> `Create` -> `FactoryPulse` -> `Factory Config`.
   - Set `factoryId = "factory-01"`.
   - Assign the 3 `LineConfig` assets.

### Step 3: Scene Setup
1. In `Assets/Scenes/MainSim.unity`:
2. Create an empty GameObject named `FactoryRoot`.
3. Add the `FactoryTopology` script.
4. Drag your `FactoryConfig` asset into the `Config` inspector field.
5. In Play mode, `FactoryTopology` auto-generates and wires the 3 lines, 12 machine nodes, and intermediate buffers.

