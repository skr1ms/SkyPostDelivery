import { useEffect, useState } from 'react';
import { goodsAPI } from '../../api';
import type { Good, CreateGoodRequest } from '../../types';
import Button from '../../components/Button/Button';
import Modal from '../../components/Modal/Modal';
import ConfirmModal from '../../components/ConfirmModal/ConfirmModal';
import { showSuccess, showError } from '../../utils/toast';
import './GoodsPage.css';

interface GoodGroup {
  name: string;
  goods: Good[];
  count: number;
  weight: number;
  height: number;
  length: number;
  width: number;
}

const GoodsPage = () => {
  const [goods, setGoods] = useState<Good[]>([]);
  const [loading, setLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false);
  const [groupToDelete, setGroupToDelete] = useState<GoodGroup | null>(null);
  const [editingGroup, setEditingGroup] = useState<GoodGroup | null>(null);
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  const [searchQuery, setSearchQuery] = useState('');
  const [formData, setFormData] = useState<CreateGoodRequest>({
    name: '',
    weight: 0,
    height: 0,
    length: 0,
    width: 0,
    quantity: 1,
  });

  useEffect(() => {
    loadGoods();
  }, []);

  const loadGoods = async () => {
    try {
      setLoading(true);
      const response = await goodsAPI.getAll();
      setGoods(response.data);
    } catch (error) {
      console.error('Failed to load goods:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (editingGroup) {
      await handleUpdateGroup();
    } else {
      setLoading(true);
      try {
        await goodsAPI.create(formData);
        showSuccess(`Создано товаров: ${formData.quantity}`);
        await loadGoods();
        closeModal();
      } catch (error: any) {
        console.error('Failed to save good:', error);
        const errorMsg = error.response?.data?.error || error.message || 'Неизвестная ошибка';
        showError(`Ошибка: ${errorMsg}`);
      } finally {
        setLoading(false);
      }
    }
  };

  const handleDeleteGroup = async (group: GoodGroup) => {
    setGroupToDelete(group);
    setIsConfirmModalOpen(true);
  };

  const confirmDelete = async () => {
    if (!groupToDelete) return;
    
    setLoading(true);
    let deletedCount = 0;
    let errorOccurred = false;

    try {
      for (const good of groupToDelete.goods) {
        try {
          await goodsAPI.delete(good.id);
          deletedCount++;
          await new Promise(resolve => setTimeout(resolve, 50));
        } catch (error: any) {
          console.error('Failed to delete good:', error);
          if (error.response?.status === 429) {
            await new Promise(resolve => setTimeout(resolve, 200));
            try {
              await goodsAPI.delete(good.id);
              deletedCount++;
            } catch (retryError) {
              errorOccurred = true;
              break;
            }
          } else {
            errorOccurred = true;
            break;
          }
        }
      }

      if (deletedCount > 0) {
        showSuccess(`Удалено товаров: ${deletedCount} из ${groupToDelete.count}`);
        await loadGoods();
      }

      if (errorOccurred && deletedCount < groupToDelete.count) {
        showError(`Удалось удалить только ${deletedCount} из ${groupToDelete.count} товаров`);
      }
    } catch (error: any) {
      console.error('Failed to delete goods:', error);
      const errorMsg = error.response?.data?.error || error.message || 'Неизвестная ошибка';
      showError(`Ошибка при удалении: ${errorMsg}`);
    } finally {
      setLoading(false);
      setIsConfirmModalOpen(false);
      setGroupToDelete(null);
    }
  };

  const handleUpdateGroup = async () => {
    if (!editingGroup) return;
    setLoading(true);
    try {
      for (const good of editingGroup.goods) {
        const { quantity, ...updateData } = formData;
        await goodsAPI.update(good.id, updateData);
      }
      showSuccess(`Обновлено товаров: ${editingGroup.count}`);
      await loadGoods();
      closeModal();
    } catch (error: any) {
      console.error('Failed to update goods:', error);
      const errorMsg = error.response?.data?.error || error.message || 'Неизвестная ошибка';
      showError(`Ошибка при обновлении: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const openModal = (group?: GoodGroup) => {
    if (group) {
      setEditingGroup(group);
      setFormData({
        name: group.name,
        weight: group.weight,
        height: group.height,
        length: group.length,
        width: group.width,
        quantity: 1,
      });
    } else {
      setEditingGroup(null);
      setFormData({ name: '', weight: 0, height: 0, length: 0, width: 0, quantity: 1 });
    }
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setEditingGroup(null);
    setFormData({ name: '', weight: 0, height: 0, length: 0, width: 0, quantity: 1 });
  };

  const groupGoods = (): GoodGroup[] => {
    const grouped = goods.reduce((acc, good) => {
      if (!acc[good.name]) {
        acc[good.name] = [];
      }
      acc[good.name].push(good);
      return acc;
    }, {} as Record<string, Good[]>);

    return Object.entries(grouped).map(([name, goods]) => ({
      name,
      goods,
      count: goods.length,
      weight: goods[0].weight,
      height: goods[0].height,
      length: goods[0].length,
      width: goods[0].width,
    }));
  };

  const getSortedAndFilteredGroups = (): GoodGroup[] => {
    let groups = groupGoods();

    if (searchQuery) {
      groups = groups.filter(group => 
        group.name.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    groups.sort((a, b) => {
      const comparison = a.name.localeCompare(b.name, 'ru');
      return sortOrder === 'asc' ? comparison : -comparison;
    });

    return groups;
  };

  return (
    <div className="goods-page">
      <div className="page-header">
        <div>
          <h1>📦 Управление Товарами</h1>
          <p>Просмотр и редактирование товаров по группам</p>
        </div>
        <Button onClick={() => openModal()}>
          ➕ Добавить товар
        </Button>
      </div>

      <div className="filters-bar">
        <div className="search-box">
          <input
            type="text"
            placeholder="🔍 Поиск по названию..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <Button 
          size="small" 
          variant="secondary" 
          onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
        >
          {sortOrder === 'asc' ? '↑ А-Я' : '↓ Я-А'}
        </Button>
      </div>

      {loading ? (
        <div className="loading">Загрузка...</div>
      ) : (
        <div className="goods-groups">
          {getSortedAndFilteredGroups().map((group, index) => (
            <div key={group.name}>
              <div className="goods-group">
                <div className="group-header">
                  <div className="group-title">
                    <h3>{group.name}</h3>
                    <span className="group-count">Количество: {group.count} шт.</span>
                  </div>
                  <div className="group-actions">
                    <Button size="small" variant="secondary" onClick={() => openModal(group)} disabled={loading}>
                      ✏️ Изменить группу
                    </Button>
                    <Button size="small" variant="danger" onClick={() => handleDeleteGroup(group)} disabled={loading}>
                      🗑️ Удалить группу
                    </Button>
                  </div>
                </div>
                <div className="group-details">
                  <div className="detail-item">
                    <strong>Вес:</strong> {group.weight} кг
                  </div>
                  <div className="detail-item">
                    <strong>Размеры:</strong> {group.length} × {group.width} × {group.height} см
                  </div>
                  <div className="detail-item">
                    <strong>В наличии:</strong> <span style={{color: group.goods[0].quantity_available > 0 ? '#4ade80' : '#f87171'}}>{group.goods[0].quantity_available} шт</span>
                  </div>
                </div>
              </div>
              {index < getSortedAndFilteredGroups().length - 1 && <div className="group-divider" />}
            </div>
          ))}
        </div>
      )}

      {goods.length === 0 && !loading && (
        <div className="empty-state">
          <p>Нет товаров. Добавьте первый товар!</p>
        </div>
      )}

      <Modal
        isOpen={isModalOpen}
        onClose={closeModal}
        title={editingGroup ? `Редактировать группу "${editingGroup.name}"` : 'Добавить товар'}
      >
        <form onSubmit={handleSubmit} className="good-form">
          <div className="form-group">
            <label>Название</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              required
              placeholder="Например: Планшет Samsung"
              disabled={!!editingGroup}
            />
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>Вес (кг)</label>
              <input
                type="number"
                step="0.01"
                value={formData.weight}
                onChange={(e) => setFormData({ ...formData, weight: parseFloat(e.target.value) })}
                required
                min="0"
              />
            </div>

            <div className="form-group">
              <label>Высота (см)</label>
              <input
                type="number"
                step="0.1"
                value={formData.height}
                onChange={(e) => setFormData({ ...formData, height: parseFloat(e.target.value) })}
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
                value={formData.length}
                onChange={(e) => setFormData({ ...formData, length: parseFloat(e.target.value) })}
                required
                min="0"
              />
            </div>

            <div className="form-group">
              <label>Ширина (см)</label>
              <input
                type="number"
                step="0.1"
                value={formData.width}
                onChange={(e) => setFormData({ ...formData, width: parseFloat(e.target.value) })}
                required
                min="0"
              />
            </div>
          </div>

          {!editingGroup && (
            <div className="form-group">
              <label>Количество</label>
              <input
                type="number"
                value={formData.quantity}
                onChange={(e) => setFormData({ ...formData, quantity: parseInt(e.target.value) || 1 })}
                required
                min="1"
              />
            </div>
          )}

          <div className="form-actions">
            <Button type="button" variant="secondary" onClick={closeModal} disabled={loading}>
              Отмена
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? '⏳ Загрузка...' : (editingGroup ? `Изменить все (${editingGroup.count} шт.)` : 'Создать')}
            </Button>
          </div>
        </form>
      </Modal>

      <ConfirmModal
        isOpen={isConfirmModalOpen}
        onClose={() => {
          setIsConfirmModalOpen(false);
          setGroupToDelete(null);
        }}
        onConfirm={confirmDelete}
        title="Подтверждение удаления"
        message={
          groupToDelete ? (
            <div>
              <p>Вы уверены, что хотите удалить все товары?</p>
              <p><strong>{groupToDelete.name}</strong></p>
              <p>Количество: <strong>{groupToDelete.count} шт.</strong></p>
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

export default GoodsPage;
