import { ReactNode } from 'react';
import Modal from '../Modal/Modal';
import Button from '../Button/Button';
import './ConfirmModal.css';

interface ConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string | ReactNode;
  confirmText?: string;
  cancelText?: string;
  variant?: 'primary' | 'secondary' | 'danger';
  loading?: boolean;
}

const ConfirmModal = ({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmText = 'OK',
  cancelText = 'Cancel',
  variant = 'danger',
  loading = false,
}: ConfirmModalProps) => {
  const handleConfirm = () => {
    onConfirm();
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={title}>
      <div className={`confirm-modal confirm-modal-${variant}`}>
        <div className="confirm-message">{message}</div>
        <div className="confirm-actions">
          <Button variant="secondary" onClick={onClose} disabled={loading}>
            {cancelText}
          </Button>
          <Button variant={variant} onClick={handleConfirm} disabled={loading}>
            {loading ? '⏳ Загрузка...' : confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  );
};

export default ConfirmModal;

