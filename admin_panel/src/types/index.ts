export interface Drone {
  id: string;
  model: string;
  ip_address: string;
  status: 'idle' | 'busy' | 'delivering' | 'error';
  battery_level: number;
  current_delivery_id?: string | null;
  speed?: number;
  error_message?: string | null;
}

export interface Good {
  id: string;
  name: string;
  weight: number;
  height: number;
  length: number;
  width: number;
  quantity_available: number;
}

export interface ParcelAutomat {
  id: string;
  city: string;
  address: string;
  number_of_cells: number;
  ip_address?: string;
  coordinates?: string;
  aruco_id: number;
  is_working: boolean;
}

export interface LockerCell {
  id: string;
  post_id: string;
  height: number;
  length: number;
  width: number;
  status: 'available' | 'reserved' | 'occupied';
}

export interface CreateDroneRequest {
  model: string;
  ip_address?: string;
  status?: string;
}

export interface CellDimensions {
  height: number;
  length: number;
  width: number;
}

export interface CreateGoodRequest {
  name: string;
  weight: number;
  height: number;
  length: number;
  width: number;
  quantity: number;
}

export interface CreateParcelAutomatRequest {
  city: string;
  address: string;
  ip_address?: string;
  coordinates?: string;
  aruco_id: number;
  number_of_cells: number;
  cells: CellDimensions[];
  is_working?: boolean;
}

export interface UpdateCellRequest {
  height: number;
  length: number;
  width: number;
}
