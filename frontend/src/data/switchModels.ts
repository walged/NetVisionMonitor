// Switch manufacturers and models database with SFP port configuration

export interface SwitchModel {
  name: string
  portCount: number      // Total number of ports
  sfpPortCount: number   // Number of SFP ports (always last N ports)
  description?: string
}

export interface Manufacturer {
  id: string
  name: string
  models: SwitchModel[]
}

export const switchManufacturers: Manufacturer[] = [
  {
    id: 'tfortis',
    name: 'TFortis',
    models: [
      // PSW-1G4F series (1 SFP + 4 FE = 5 ports)
      { name: 'PSW-1G4F', portCount: 5, sfpPortCount: 1, description: '4 PoE + 1 SFP' },
      { name: 'PSW-1G4F-Box', portCount: 5, sfpPortCount: 1, description: '4 PoE + 1 SFP, Box' },
      { name: 'PSW-1G4F-Ex', portCount: 5, sfpPortCount: 1, description: '4 PoE + 1 SFP, Ex' },
      { name: 'PSW-1G4F-UPS', portCount: 5, sfpPortCount: 1, description: '4 PoE + 1 SFP, UPS' },
      // PSW-2G4F series (2 SFP + 4 FE = 6 ports)
      { name: 'PSW-2G4F', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP' },
      { name: 'PSW-2G4F-Box', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP, Box' },
      { name: 'PSW-2G4F-Ex', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP, Ex' },
      { name: 'PSW-2G4F-UPS', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP, UPS' },
      // PSW-2G-Plus series (2 SFP + 2 FE = 4 ports)
      { name: 'PSW-2G+', portCount: 4, sfpPortCount: 2, description: '2 PoE + 2 SFP' },
      { name: 'PSW-2G+-Box', portCount: 4, sfpPortCount: 2, description: '2 PoE + 2 SFP, Box' },
      { name: 'PSW-2G+-Ex', portCount: 4, sfpPortCount: 2, description: '2 PoE + 2 SFP, Ex' },
      { name: 'PSW-2G+-UPS-Box', portCount: 4, sfpPortCount: 2, description: '2 PoE + 2 SFP, UPS' },
      // PSW-2G2F-Plus-UPS (2 SFP + 2 FE = 4 ports)
      { name: 'PSW-2G2F+-UPS', portCount: 4, sfpPortCount: 2, description: '2 PoE + 2 SFP, UPS' },
      // PSW-2G6F-Plus series (2 SFP + 6 FE = 8 ports)
      { name: 'PSW-2G6F+', portCount: 8, sfpPortCount: 2, description: '6 PoE + 2 SFP' },
      { name: 'PSW-2G6F+-Box', portCount: 8, sfpPortCount: 2, description: '6 PoE + 2 SFP, Box' },
      { name: 'PSW-2G6F+-UPS-Box', portCount: 8, sfpPortCount: 2, description: '6 PoE + 2 SFP, UPS' },
      // PSW-2G8F-Plus series (2 SFP + 8 FE = 10 ports)
      { name: 'PSW-2G8F+', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP' },
      { name: 'PSW-2G8F+-Box', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP, Box' },
      { name: 'PSW-2G8F+-UPS-Box', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP, UPS' },
      // PSW-2G series (2 SFP + 3 PoE = 5 ports)
      { name: 'PSW-2G', portCount: 5, sfpPortCount: 2, description: '3 PoE + 2 SFP' },
    ],
  },
  {
    id: 'cisco',
    name: 'Cisco',
    models: [
      { name: 'SG250-08', portCount: 8, sfpPortCount: 0, description: '8-Port Gigabit' },
      { name: 'SG250-10P', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP' },
      { name: 'SG250-18', portCount: 18, sfpPortCount: 2, description: '16 ports + 2 SFP' },
      { name: 'SG250-26', portCount: 26, sfpPortCount: 2, description: '24 ports + 2 SFP' },
      { name: 'SG250-50', portCount: 50, sfpPortCount: 2, description: '48 ports + 2 SFP' },
      { name: 'SG350-10', portCount: 10, sfpPortCount: 2, description: '8 ports + 2 SFP' },
      { name: 'SG350-10P', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP' },
      { name: 'SG350-28', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'SG350-52', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
      { name: 'CBS250-8T', portCount: 8, sfpPortCount: 0, description: '8-Port Business' },
      { name: 'CBS250-24T', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'CBS250-48T', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
      { name: 'CBS350-8T', portCount: 10, sfpPortCount: 2, description: '8 ports + 2 SFP' },
      { name: 'CBS350-24T', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'CBS350-48T', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
    ],
  },
  {
    id: 'dlink',
    name: 'D-Link',
    models: [
      { name: 'DGS-1100-05', portCount: 5, sfpPortCount: 0, description: '5-Port Gigabit' },
      { name: 'DGS-1100-08', portCount: 8, sfpPortCount: 0, description: '8-Port Gigabit' },
      { name: 'DGS-1100-16', portCount: 16, sfpPortCount: 0, description: '16-Port Gigabit' },
      { name: 'DGS-1100-24', portCount: 24, sfpPortCount: 0, description: '24-Port Gigabit' },
      { name: 'DGS-1210-10', portCount: 10, sfpPortCount: 2, description: '8 ports + 2 SFP' },
      { name: 'DGS-1210-20', portCount: 20, sfpPortCount: 4, description: '16 ports + 4 SFP' },
      { name: 'DGS-1210-28', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'DGS-1210-52', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
      { name: 'DGS-1510-20', portCount: 20, sfpPortCount: 4, description: '16 ports + 4 SFP' },
      { name: 'DGS-1510-28', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'DGS-1510-52', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
    ],
  },
  {
    id: 'tplink',
    name: 'TP-Link',
    models: [
      { name: 'TL-SG108', portCount: 8, sfpPortCount: 0, description: '8-Port Gigabit' },
      { name: 'TL-SG116', portCount: 16, sfpPortCount: 0, description: '16-Port Gigabit' },
      { name: 'TL-SG1024', portCount: 24, sfpPortCount: 0, description: '24-Port Gigabit' },
      { name: 'TL-SG2008', portCount: 8, sfpPortCount: 0, description: '8-Port Smart' },
      { name: 'TL-SG2016', portCount: 16, sfpPortCount: 0, description: '16-Port Smart' },
      { name: 'TL-SG2024', portCount: 24, sfpPortCount: 0, description: '24-Port Smart' },
      { name: 'TL-SG2210P', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP' },
      { name: 'TL-SG2428P', portCount: 28, sfpPortCount: 4, description: '24 PoE + 4 SFP' },
      { name: 'TL-SG3210', portCount: 10, sfpPortCount: 2, description: '8 ports + 2 SFP' },
      { name: 'TL-SG3428', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'TL-SG3452', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
    ],
  },
  {
    id: 'mikrotik',
    name: 'MikroTik',
    models: [
      { name: 'CSS106-5G-1S', portCount: 6, sfpPortCount: 1, description: '5 ports + 1 SFP' },
      { name: 'CSS610-8G-2S+', portCount: 10, sfpPortCount: 2, description: '8 ports + 2 SFP+' },
      { name: 'CRS112-8G-4S', portCount: 12, sfpPortCount: 4, description: '8 ports + 4 SFP' },
      { name: 'CRS112-8P-4S', portCount: 12, sfpPortCount: 4, description: '8 PoE + 4 SFP' },
      { name: 'CRS328-24P-4S+', portCount: 28, sfpPortCount: 4, description: '24 PoE + 4 SFP+' },
      { name: 'CRS326-24G-2S+', portCount: 26, sfpPortCount: 2, description: '24 ports + 2 SFP+' },
      { name: 'CRS354-48G-4S+', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP+' },
      { name: 'CRS354-48P-4S+', portCount: 52, sfpPortCount: 4, description: '48 PoE + 4 SFP+' },
    ],
  },
  {
    id: 'hp',
    name: 'HPE/Aruba',
    models: [
      { name: 'Aruba 1930 8G', portCount: 10, sfpPortCount: 2, description: '8 ports + 2 SFP' },
      { name: 'Aruba 1930 24G', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'Aruba 1930 48G', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
      { name: 'Aruba 1960 24G', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'Aruba 1960 48G', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
      { name: 'Aruba 2530-8G', portCount: 10, sfpPortCount: 2, description: '8 ports + 2 SFP' },
      { name: 'Aruba 2530-24G', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'Aruba 2530-48G', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
    ],
  },
  {
    id: 'ubiquiti',
    name: 'Ubiquiti',
    models: [
      { name: 'USW-Lite-8-PoE', portCount: 8, sfpPortCount: 0, description: '8-Port Lite PoE' },
      { name: 'USW-Lite-16-PoE', portCount: 16, sfpPortCount: 0, description: '16-Port Lite PoE' },
      { name: 'US-8-150W', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP' },
      { name: 'US-16-150W', portCount: 18, sfpPortCount: 2, description: '16 PoE + 2 SFP' },
      { name: 'USW-24', portCount: 26, sfpPortCount: 2, description: '24 ports + 2 SFP' },
      { name: 'USW-24-PoE', portCount: 26, sfpPortCount: 2, description: '24 PoE + 2 SFP' },
      { name: 'USW-48', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
      { name: 'USW-48-PoE', portCount: 52, sfpPortCount: 4, description: '48 PoE + 4 SFP' },
      { name: 'USW-Pro-24', portCount: 26, sfpPortCount: 2, description: '24 ports + 2 SFP+' },
      { name: 'USW-Pro-24-PoE', portCount: 26, sfpPortCount: 2, description: '24 PoE + 2 SFP+' },
      { name: 'USW-Pro-48', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP+' },
      { name: 'USW-Pro-48-PoE', portCount: 52, sfpPortCount: 4, description: '48 PoE + 4 SFP+' },
    ],
  },
  {
    id: 'ltv',
    name: 'LTV',
    models: [
      // 2-Series (неуправляемые PoE коммутаторы)
      { name: 'LTV-2S04F2U-P', portCount: 6, sfpPortCount: 0, description: '4 PoE + 2 Uplink' },
      { name: 'LTV-2S08F2U-P', portCount: 10, sfpPortCount: 0, description: '8 PoE + 2 Uplink' },
      { name: 'LTV-2S16F2C-P', portCount: 18, sfpPortCount: 2, description: '16 PoE + 2 Combo SFP' },
      { name: 'LTV-2S24F2C-P', portCount: 26, sfpPortCount: 2, description: '24 PoE + 2 Combo SFP' },
      { name: 'LTV-2S48F2C-P', portCount: 50, sfpPortCount: 2, description: '48 PoE + 2 Combo SFP' },
      // 2-Series Gigabit
      { name: 'LTV-2S24G4C', portCount: 28, sfpPortCount: 4, description: '24 GE + 4 Combo SFP' },
      // 2-Series Outdoor (уличные)
      { name: 'LTV-2SE04G2S-HP', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP, уличный' },
      { name: 'LTV-2SE04G2S-UHP', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP, уличный' },
      { name: 'LTV-2SE04G2S-MHP', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP, уличный' },
      { name: 'LTV-2SE04G2S-MUHP', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP, уличный' },
      { name: 'LTV-2SE08G2S-HP', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP, уличный' },
      { name: 'LTV-2SE08G2S-UHP', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP, уличный' },
      { name: 'LTV-2SE08G2S-MHP', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP, уличный' },
      { name: 'LTV-2SE08G2S-MUHP', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP, уличный' },
      // 2-Series Industrial (промышленные)
      { name: 'LTV-2SI04G2S-MP', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP, промышленный' },
      { name: 'LTV-2SI08G2S-MP', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP, промышленный' },
      { name: 'LTV-2SI08G2S-P', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP, промышленный' },
      // 3-Series (управляемые L2+ коммутаторы)
      { name: 'LTV-3S08G4C-MP', portCount: 12, sfpPortCount: 4, description: '8 PoE + 4 Combo, L2+' },
      { name: 'LTV-3S16G2H-MP', portCount: 18, sfpPortCount: 2, description: '16 PoE + 2 SFP, L2+' },
      { name: 'LTV-3S24G4C-MP', portCount: 28, sfpPortCount: 4, description: '24 PoE + 4 Combo, L2+' },
      // 3-Series Aggregation (агрегация)
      { name: 'LTV-3S08G18S-M', portCount: 26, sfpPortCount: 18, description: '8 GE + 18 SFP, агрегация' },
      { name: 'LTV-3S24G4S-M', portCount: 28, sfpPortCount: 4, description: '24 GE + 4 SFP, агрегация' },
      // 3-Series Outdoor (уличные управляемые)
      { name: 'LTV-3SE04G2S-MUBP', portCount: 6, sfpPortCount: 2, description: '4 PoE + 2 SFP, уличный L2+' },
      { name: 'LTV-3SE08G2S-MUBP', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP, уличный L2+' },
      // Старая серия NSF
      { name: 'LTV-NSF-0804', portCount: 8, sfpPortCount: 0, description: '8 PoE (устаревшая)' },
      { name: 'LTV-NSF-1610', portCount: 18, sfpPortCount: 2, description: '16 PoE + 2 SFP (устаревшая)' },
      { name: 'LTV-NSF-2724', portCount: 26, sfpPortCount: 2, description: '24 PoE + 2 SFP (устаревшая)' },
    ],
  },
  {
    id: 'hikvision',
    name: 'Hikvision',
    models: [
      { name: 'DS-3E0105P-E', portCount: 5, sfpPortCount: 0, description: '4 PoE + 1 Uplink' },
      { name: 'DS-3E0109P-E', portCount: 9, sfpPortCount: 0, description: '8 PoE + 1 Uplink' },
      { name: 'DS-3E0318P-E', portCount: 18, sfpPortCount: 2, description: '16 PoE + 2 SFP' },
      { name: 'DS-3E0326P-E', portCount: 26, sfpPortCount: 2, description: '24 PoE + 2 SFP' },
      { name: 'DS-3E1310P-SI', portCount: 10, sfpPortCount: 2, description: '8 PoE + 2 SFP' },
      { name: 'DS-3E1318P-SI', portCount: 18, sfpPortCount: 2, description: '16 PoE + 2 SFP' },
      { name: 'DS-3E1326P-SI', portCount: 26, sfpPortCount: 2, description: '24 PoE + 2 SFP' },
      { name: 'DS-3E1552P-SI', portCount: 52, sfpPortCount: 4, description: '48 PoE + 4 SFP' },
    ],
  },
  {
    id: 'dahua',
    name: 'Dahua',
    models: [
      { name: 'PFS3005-4ET-60', portCount: 5, sfpPortCount: 0, description: '4 PoE + 1 Uplink' },
      { name: 'PFS3010-8ET-96', portCount: 10, sfpPortCount: 0, description: '8 PoE + 2 Uplink' },
      { name: 'PFS4218-16ET-240', portCount: 18, sfpPortCount: 2, description: '16 PoE + 2 SFP' },
      { name: 'PFS4226-24ET-240', portCount: 26, sfpPortCount: 2, description: '24 PoE + 2 SFP' },
      { name: 'PFS4226-24ET-360', portCount: 26, sfpPortCount: 2, description: '24 PoE + 2 SFP' },
      { name: 'PFS4428-24GT-370', portCount: 28, sfpPortCount: 4, description: '24 PoE + 4 SFP' },
    ],
  },
  {
    id: 'netgear',
    name: 'NETGEAR',
    models: [
      { name: 'GS108', portCount: 8, sfpPortCount: 0, description: '8-Port Gigabit' },
      { name: 'GS116', portCount: 16, sfpPortCount: 0, description: '16-Port Gigabit' },
      { name: 'GS308E', portCount: 8, sfpPortCount: 0, description: '8-Port Plus' },
      { name: 'GS316EP', portCount: 16, sfpPortCount: 0, description: '16-Port Plus PoE+' },
      { name: 'GS724T', portCount: 26, sfpPortCount: 2, description: '24 ports + 2 SFP' },
      { name: 'GS748T', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
      { name: 'M4250-10G2F-PoE+', portCount: 12, sfpPortCount: 2, description: '10 PoE + 2 SFP' },
      { name: 'M4250-26G4F-PoE+', portCount: 30, sfpPortCount: 4, description: '26 PoE + 4 SFP' },
    ],
  },
  {
    id: 'zyxel',
    name: 'ZyXEL',
    models: [
      { name: 'GS1200-5', portCount: 5, sfpPortCount: 0, description: '5-Port Web Managed' },
      { name: 'GS1200-8', portCount: 8, sfpPortCount: 0, description: '8-Port Web Managed' },
      { name: 'GS1900-8', portCount: 10, sfpPortCount: 2, description: '8 ports + 2 SFP' },
      { name: 'GS1900-24', portCount: 26, sfpPortCount: 2, description: '24 ports + 2 SFP' },
      { name: 'GS1900-48', portCount: 50, sfpPortCount: 2, description: '48 ports + 2 SFP' },
      { name: 'XGS1930-28', portCount: 28, sfpPortCount: 4, description: '24 ports + 4 SFP' },
      { name: 'XGS1930-52', portCount: 52, sfpPortCount: 4, description: '48 ports + 4 SFP' },
    ],
  },
  {
    id: 'other',
    name: 'Другой',
    models: [
      { name: '4 порта', portCount: 4, sfpPortCount: 0, description: '4 copper' },
      { name: '4+1 SFP', portCount: 5, sfpPortCount: 1, description: '4 copper + 1 SFP' },
      { name: '4+2 SFP', portCount: 6, sfpPortCount: 2, description: '4 copper + 2 SFP' },
      { name: '8 портов', portCount: 8, sfpPortCount: 0, description: '8 copper' },
      { name: '8+2 SFP', portCount: 10, sfpPortCount: 2, description: '8 copper + 2 SFP' },
      { name: '16 портов', portCount: 16, sfpPortCount: 0, description: '16 copper' },
      { name: '16+2 SFP', portCount: 18, sfpPortCount: 2, description: '16 copper + 2 SFP' },
      { name: '24 порта', portCount: 24, sfpPortCount: 0, description: '24 copper' },
      { name: '24+2 SFP', portCount: 26, sfpPortCount: 2, description: '24 copper + 2 SFP' },
      { name: '24+4 SFP', portCount: 28, sfpPortCount: 4, description: '24 copper + 4 SFP' },
      { name: '48 портов', portCount: 48, sfpPortCount: 0, description: '48 copper' },
      { name: '48+4 SFP', portCount: 52, sfpPortCount: 4, description: '48 copper + 4 SFP' },
    ],
  },
]

// Helper function to get models by manufacturer
export function getModelsByManufacturer(manufacturerId: string): SwitchModel[] {
  const manufacturer = switchManufacturers.find((m) => m.id === manufacturerId)
  return manufacturer?.models || []
}

// Helper function to get port count for a specific model
export function getPortCountForModel(manufacturerId: string, modelName: string): number | undefined {
  const models = getModelsByManufacturer(manufacturerId)
  const model = models.find((m) => m.name === modelName)
  return model?.portCount
}

// Helper function to get SFP port count for a specific model
export function getSfpPortCountForModel(manufacturerId: string, modelName: string): number {
  const models = getModelsByManufacturer(manufacturerId)
  const model = models.find((m) => m.name === modelName)
  return model?.sfpPortCount || 0
}

// Helper function to get full model config
export function getModelConfig(manufacturerId: string, modelName: string): SwitchModel | undefined {
  const models = getModelsByManufacturer(manufacturerId)
  return models.find((m) => m.name === modelName)
}
