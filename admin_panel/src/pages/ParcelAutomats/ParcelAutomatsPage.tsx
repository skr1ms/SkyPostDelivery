import { useEffect, useState } from 'react';
import { parcelAutomatsAPI } from '../../api';
import type { ParcelAutomat, CreateParcelAutomatRequest, LockerCell, UpdateCellRequest, CellDimensions } from '../../types';
import Button from '../../components/Button/Button';
import Modal from '../../components/Modal/Modal';
import ConfirmModal from '../../components/ConfirmModal/ConfirmModal';
import { showSuccess, showError } from '../../utils/toast';
import './ParcelAutomatsPage.css';

const ParcelAutomatsPage = () => {
  const [automats, setAutomats] = useState<ParcelAutomat[]>([]);
  const [loading, setLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isCellsModalOpen, setIsCellsModalOpen] = useState(false);
  const [isStatusModalOpen, setIsStatusModalOpen] = useState(false);
  const [isEditCellModalOpen, setIsEditCellModalOpen] = useState(false);
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false);
  const [automatToDelete, setAutomatToDelete] = useState<ParcelAutomat | null>(null);
  const [editingAutomat, setEditingAutomat] = useState<ParcelAutomat | null>(null);
  const [editingCell, setEditingCell] = useState<LockerCell | null>(null);
  const [selectedAutomatId, setSelectedAutomatId] = useState<string>('');
  const [selectedAutomatCells, setSelectedAutomatCells] = useState<LockerCell[]>([]);
  const [formData, setFormData] = useState<CreateParcelAutomatRequest>({
    city: '',
    address: '',
    ip_address: '',
    coordinates: '',
    aruco_id: 0,
    number_of_cells: 0,
    cells: [],
    is_working: true,
  });
  const [editFormData, setEditFormData] = useState({
    city: '',
    address: '',
    ip_address: '',
    coordinates: '',
  });
  const [cellFormData, setCellFormData] = useState<UpdateCellRequest>({
    height: 0,
    length: 0,
    width: 0,
  });

  useEffect(() => {
    loadAutomats();
  }, []);

  const loadAutomats = async () => {
    try {
      setLoading(true);
      const response = await parcelAutomatsAPI.getAll();
      
      if (response.data && Array.isArray(response.data)) {
        const normalizedAutomats = response.data.map((automat: any) => ({
          id: automat.id || automat.ID,
          city: automat.city || automat.City,
          address: automat.address || automat.Address,
          number_of_cells: automat.number_of_cells || automat.NumberOfCells,
          ip_address: automat.ip_address || automat.IPAddress || automat.IpAddress,
          coordinates: automat.coordinates || automat.Coordinates,
          aruco_id: automat.aruco_id || automat.ArucoID || 0,
          is_working: automat.is_working !== undefined ? automat.is_working : automat.IsWorking,
        }));
        setAutomats(normalizedAutomats);
      } else {
        setAutomats([]);
      }
    } catch (error) {
      console.error('Failed to load automats:', error);
      setAutomats([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      await parcelAutomatsAPI.create(formData);
      showSuccess('Постамат успешно создан!');
      await loadAutomats();
      closeModal();
    } catch (error: any) {
      console.error('Failed to save automat:', error);
      const errorMsg = error.response?.data?.error || error.message || 'Неизвестная ошибка';
      showError(`Ошибка: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleStatusChange = async () => {
    if (!editingAutomat) return;
    setLoading(true);
    try {
      await parcelAutomatsAPI.updateStatus(editingAutomat.id, !editingAutomat.is_working);
      showSuccess('Статус постамата успешно обновлен!');
      await loadAutomats();
      setIsStatusModalOpen(false);
      setEditingAutomat(null);
    } catch (error: any) {
      console.error('Failed to update status:', error);
      const errorMsg = error.response?.data?.error || error.message || 'Неизвестная ошибка';
      showError(`Ошибка: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (automat: ParcelAutomat) => {
    setAutomatToDelete(automat);
    setIsConfirmModalOpen(true);
  };

  const confirmDelete = async () => {
    if (!automatToDelete) return;
    setLoading(true);
    try {
      await parcelAutomatsAPI.delete(automatToDelete.id);
      showSuccess('Постамат успешно удален!');
      await loadAutomats();
    } catch (error: any) {
      console.error('Failed to delete automat:', error);
      const errorMsg = error.response?.data?.error || error.message || 'Неизвестная ошибка';
      showError(`Ошибка при удалении: ${errorMsg}`);
    } finally {
      setLoading(false);
      setIsConfirmModalOpen(false);
      setAutomatToDelete(null);
    }
  };

  const openModal = () => {
    setEditingAutomat(null);
    setFormData({ city: '', address: '', ip_address: '', coordinates: '', aruco_id: 0, number_of_cells: 0, cells: [], is_working: true });
    setIsModalOpen(true);
  };

  const openEditModal = (automat: ParcelAutomat) => {
    setEditingAutomat(automat);
    setEditFormData({
      city: automat.city,
      address: automat.address,
      ip_address: automat.ip_address || '',
      coordinates: automat.coordinates || '',
    });
    setIsEditModalOpen(true);
  };

  const closeEditModal = () => {
    setIsEditModalOpen(false);
    setEditingAutomat(null);
    setEditFormData({ city: '', address: '', ip_address: '', coordinates: '' });
  };

  const handleEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingAutomat) return;
    setLoading(true);
    
    try {
      await parcelAutomatsAPI.update(editingAutomat.id, editFormData);
      showSuccess('Постамат успешно обновлен!');
      await loadAutomats();
      closeEditModal();
    } catch (error: any) {
      console.error('Failed to update automat:', error);
      const errorMsg = error.response?.data?.error || error.message || 'Неизвестная ошибка';
      showError(`Ошибка: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleCellsCountChange = (count: number) => {
    const newCells: CellDimensions[] = [];
    for (let i = 0; i < count; i++) {
      if (formData.cells[i]) {
        newCells.push(formData.cells[i]);
      } else {
        newCells.push({ height: 40.0, length: 40.0, width: 40.0 });
      }
    }
    setFormData({ ...formData, number_of_cells: count, cells: newCells });
  };

  const updateCellDimension = (index: number, field: keyof CellDimensions, value: number) => {
    const newCells = [...formData.cells];
    newCells[index] = { ...newCells[index], [field]: value };
    setFormData({ ...formData, cells: newCells });
  };

  const openStatusModal = (automat: ParcelAutomat) => {
    setEditingAutomat(automat);
    setIsStatusModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setEditingAutomat(null);
    setFormData({ city: '', address: '', aruco_id: 0, number_of_cells: 0, cells: [], is_working: true });
  };

  const viewCells = async (automat: ParcelAutomat) => {
    try {
      const response = await parcelAutomatsAPI.getCells(automat.id);
      setSelectedAutomatCells(response.data);
      setSelectedAutomatId(automat.id);
      setIsCellsModalOpen(true);
    } catch (error) {
      console.error('Failed to load cells:', error);
    }
  };

  const openEditCellModal = (cell: LockerCell) => {
    setEditingCell(cell);
    setCellFormData({
      height: cell.height,
      length: cell.length,
      width: cell.width,
    });
    setIsEditCellModalOpen(true);
  };

  const handleCellUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingCell || !selectedAutomatId) return;
    setLoading(true);

    try {
      await parcelAutomatsAPI.updateCell(selectedAutomatId, editingCell.id, cellFormData);
      showSuccess('Размеры ячейки успешно обновлены!');
      // Reload cells
      const response = await parcelAutomatsAPI.getCells(selectedAutomatId);
      setSelectedAutomatCells(response.data);
      setIsEditCellModalOpen(false);
      setEditingCell(null);
    } catch (error: any) {
      console.error('Failed to update cell:', error);
      const errorMsg = error.response?.data?.error || error.message || 'Неизвестная ошибка';
      showError(`Ошибка: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const getCellStatusBadge = (status: string) => {
    const badges: Record<string, { label: string; color: string }> = {
      available: { label: 'Свободна', color: '#00d4aa' },
      reserved: { label: 'Забронирована', color: '#ffd803' },
      occupied: { label: 'Занята', color: '#ff6b9d' },
    };
    const badge = badges[status] || { label: status, color: '#a7a9be' };
    return (
      <span className="status-badge" style={{ '--badge-color': badge.color } as React.CSSProperties}>
        {badge.label}
      </span>
    );
  };

  return (
    <div className="automats-page">
      <div className="page-header">
        <div>
          <h1>🏪 Управление Постаматами</h1>
          <p>Просмотр и редактирование постаматов</p>
        </div>
        <Button onClick={() => openModal()}>
          ➕ Добавить постамат
        </Button>
      </div>

      {loading ? (
        <div className="loading">Загрузка...</div>
      ) : automats.length > 0 ? (
        <div className="automats-grid">
          {automats.map((automat) => (
            <div key={automat.id} className="automat-card">
              <div className="automat-header">
                <h3>{automat.city}</h3>
                <span className={`working-badge ${automat.is_working ? 'working' : 'not-working'}`}>
                  {automat.is_working ? '✓ Работает' : '✕ Не работает'}
                </span>
              </div>
              <p className="automat-address">📍 {automat.address}</p>
              <div className="automat-info">
                <div><strong>ID:</strong> {automat.id ? automat.id.substring(0, 8) : 'N/A'}...</div>
                <div><strong>Ячеек:</strong> {automat.number_of_cells}</div>
              </div>
              <div className="automat-actions">
                <Button size="small" variant="secondary" onClick={() => viewCells(automat)} disabled={loading}>
                  ℹ️ Сведения
                </Button>
                <Button size="small" variant="primary" onClick={() => openEditModal(automat)} disabled={loading}>
                  ✏️ Изменить
                </Button>
                <Button size="small" variant="secondary" onClick={() => openStatusModal(automat)} disabled={loading}>
                  {automat.is_working ? '⏸️ Выключить' : '▶️ Включить'}
                </Button>
                <Button size="small" variant="danger" onClick={() => handleDelete(automat)} disabled={loading}>
                  🗑️
                </Button>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="empty-state">
          <p>Нет постаматов. Добавьте первый постамат!</p>
        </div>
      )}

      <Modal
        isOpen={isModalOpen}
        onClose={closeModal}
        title="Добавить постамат"
      >
        <form onSubmit={handleSubmit} className="automat-form">
          <div className="form-group">
            <label>Город</label>
            <input
              type="text"
              value={formData.city}
              onChange={(e) => setFormData({ ...formData, city: e.target.value })}
              required
              placeholder="Например: Екатеринбург"
            />
          </div>

                    <div className="form-group">
            <label>Адрес</label>
            <input
              type="text"
              value={formData.address}
              onChange={(e) => setFormData({ ...formData, address: e.target.value })}
              required
              placeholder="Например: ул. Ленина, 5"
            />
          </div>

          <div className="form-group">
            <label>IP-адрес</label>
            <input
              type="text"
              value={formData.ip_address}
              onChange={(e) => setFormData({ ...formData, ip_address: e.target.value })}
              placeholder="Например: 192.168.1.100"
            />
          </div>

          <div className="form-group">
            <label>Координаты</label>
            <input
              type="text"
              value={formData.coordinates}
              onChange={(e) => setFormData({ ...formData, coordinates: e.target.value })}
              placeholder="Например: 56.8389,60.6057"
            />
          </div>

          <div className="form-group">
            <label>ArUco ID маркера</label>
            <input
              type="number"
              value={formData.aruco_id || ''}
              onChange={(e) => setFormData({ ...formData, aruco_id: parseInt(e.target.value) || 0 })}
              required
              min="0"
              placeholder="Например: 42 (допустимо 0-1023)"
            />
            <small style={{ color: '#a7a9be', marginTop: '4px', display: 'block' }}>
              Примеры популярных ArUco меток: 0, 1, 42, 100, 255, 500, 1000
            </small>
          </div>

          <div className="form-group">
            <label>Количество ячеек</label>
            <input
              type="number"
              value={formData.number_of_cells || ''}
              onChange={(e) => handleCellsCountChange(parseInt(e.target.value) || 0)}
              required
              min="1"
              max="100"
              placeholder="Например: 12"
            />
          </div>

          {formData.number_of_cells > 0 && (
            <div className="cells-config">
              <h3>Конфигурация ячеек</h3>
              <div className="cells-list-config">
                {formData.cells.map((cell, index) => (
                  <div key={index} className="cell-config-item">
                    <h4>Ячейка #{index + 1}</h4>
                    <div className="form-row">
                      <div className="form-group">
                        <label>Высота (см)</label>
                        <input
                          type="number"
                          step="0.1"
                          value={cell.height}
                          onChange={(e) => updateCellDimension(index, 'height', parseFloat(e.target.value) || 0)}
                          required
                          min="0"
                        />
                      </div>
                    </div>
                    <div className="form-row">
                      <div className="form-group">
                        <label>Длина (см)</label>
                        <input
                          type="number"
                          step="0.1"
                          value={cell.length}
                          onChange={(e) => updateCellDimension(index, 'length', parseFloat(e.target.value) || 0)}
                          required
                          min="0"
                        />
                      </div>
                      <div className="form-group">
                        <label>Ширина (см)</label>
                        <input
                          type="number"
                          step="0.1"
                          value={cell.width}
                          onChange={(e) => updateCellDimension(index, 'width', parseFloat(e.target.value) || 0)}
                          required
                          min="0"
                        />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="form-actions">
            <Button type="button" variant="secondary" onClick={closeModal} disabled={loading}>
              Отмена
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? '⏳ Загрузка...' : 'Создать'}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isEditModalOpen}
        onClose={closeEditModal}
        title="Редактировать постамат"
      >
        <form onSubmit={handleEdit} className="automat-form">
          <div className="form-group">
            <label>Город</label>
            <input
              type="text"
              value={editFormData.city}
              onChange={(e) => setEditFormData({ ...editFormData, city: e.target.value })}
              required
              placeholder="Например: Екатеринбург"
            />
          </div>

          <div className="form-group">
            <label>Адрес</label>
            <input
              type="text"
              value={editFormData.address}
              onChange={(e) => setEditFormData({ ...editFormData, address: e.target.value })}
              required
              placeholder="Например: ул. Ленина, 1"
            />
          </div>

          <div className="form-group">
            <label>IP адрес</label>
            <input
              type="text"
              value={editFormData.ip_address}
              onChange={(e) => setEditFormData({ ...editFormData, ip_address: e.target.value })}
              placeholder="Например: 192.168.1.100"
            />
          </div>

          <div className="form-group">
            <label>Координаты</label>
            <input
              type="text"
              value={editFormData.coordinates}
              onChange={(e) => setEditFormData({ ...editFormData, coordinates: e.target.value })}
              placeholder="Например: 56.8389,60.6057"
            />
          </div>

          <div className="form-actions">
            <Button type="button" variant="secondary" onClick={closeEditModal} disabled={loading}>
              Отмена
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? '⏳ Загрузка...' : 'Сохранить'}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isStatusModalOpen}
        onClose={() => setIsStatusModalOpen(false)}
        title="Изменить статус постамата"
      >
        <div style={{ padding: '20px' }}>
          <p>
            Вы уверены, что хотите {editingAutomat?.is_working ? 'выключить' : 'включить'} постамат?
          </p>
          <div className="form-actions" style={{ marginTop: '20px' }}>
            <Button type="button" variant="secondary" onClick={() => setIsStatusModalOpen(false)} disabled={loading}>
              Отмена
            </Button>
            <Button onClick={handleStatusChange} disabled={loading}>
              {loading ? '⏳ Загрузка...' : 'Подтвердить'}
            </Button>
          </div>
        </div>
      </Modal>

      <Modal
        isOpen={isCellsModalOpen}
        onClose={() => setIsCellsModalOpen(false)}
        title="Ячейки постамата"
      >
        <div className="cells-list">
          {selectedAutomatCells.map((cell, index) => (
            <div key={cell.id} className="cell-item">
              <div className="cell-header">
                <strong>Ячейка #{index + 1}</strong>
                {getCellStatusBadge(cell.status)}
              </div>
              <div className="cell-dimensions">
                <span>📏 Д: {cell.length} см × Ш: {cell.width} см × В: {cell.height} см</span>
              </div>
              {cell.status === 'available' && (
                <Button 
                  size="small" 
                  variant="secondary" 
                  onClick={() => openEditCellModal(cell)}
                  disabled={loading}
                  style={{ marginTop: '8px' }}
                >
                  ✏️ Редактировать размеры
                </Button>
              )}
            </div>
          ))}
          {selectedAutomatCells.length === 0 && (
            <p className="empty-cells">Нет данных о ячейках</p>
          )}
        </div>
      </Modal>

      <Modal
        isOpen={isEditCellModalOpen}
        onClose={() => {
          setIsEditCellModalOpen(false);
          setEditingCell(null);
        }}
        title="Редактировать размеры ячейки"
      >
        <form onSubmit={handleCellUpdate} className="automat-form">
          <div className="form-group">
            <label>Высота (см)</label>
            <input
              type="number"
              step="0.1"
              value={cellFormData.height}
              onChange={(e) => setCellFormData({ ...cellFormData, height: parseFloat(e.target.value) })}
              required
              min="0"
            />
          </div>

          <div className="form-group">
            <label>Длина (см)</label>
            <input
              type="number"
              step="0.1"
              value={cellFormData.length}
              onChange={(e) => setCellFormData({ ...cellFormData, length: parseFloat(e.target.value) })}
              required
              min="0"
            />
          </div>

          <div className="form-group">
            <label>Ширина (см)</label>
            <input
              type="number"
              step="0.1"
              value={cellFormData.width}
              onChange={(e) => setCellFormData({ ...cellFormData, width: parseFloat(e.target.value) })}
              required
              min="0"
            />
          </div>

          <div className="form-actions">
            <Button 
              type="button" 
              variant="secondary" 
              onClick={() => {
                setIsEditCellModalOpen(false);
                setEditingCell(null);
              }}
              disabled={loading}
            >
              Отмена
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? '⏳ Загрузка...' : 'Сохранить'}
            </Button>
          </div>
        </form>
      </Modal>

      <ConfirmModal
        isOpen={isConfirmModalOpen}
        onClose={() => {
          setIsConfirmModalOpen(false);
          setAutomatToDelete(null);
        }}
        onConfirm={confirmDelete}
        title="Подтверждение удаления"
        message={
          automatToDelete ? (
            <div>
              <p>Вы уверены, что хотите удалить постамат?</p>
              <p><strong>{automatToDelete.city}, {automatToDelete.address}</strong></p>
              <p>Количество ячеек: <strong>{automatToDelete.number_of_cells}</strong></p>
            </div>
          ) : ''
        }
        confirmText="Удалить"
        cancelText="Отмена"
        variant="danger"
        loading={loading}
      />
    </div>
  );
};

export default ParcelAutomatsPage;
