import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { 
  Bell, 
  Check, 
  CheckCheck,
  ChevronLeft,
  Loader2,
  AlertCircle,
  Trash2,
  Flag,
  AlertTriangle,
  Megaphone,
  Settings,
  ExternalLink,
} from 'lucide-react';
import { notificationsApi } from '../../services/flagflash-api';
import { ConfirmDeleteModal } from '../../components';
import type { Notification, NotificationType } from '../../types/flagflash';

const TYPE_ICONS: Record<NotificationType, React.ReactNode> = {
  flag_change: <Flag size={16} className="text-accent-blue" />,
  alert: <AlertTriangle size={16} className="text-accent-amber" />,
  announcement: <Megaphone size={16} className="text-accent-purple" />,
  system: <Settings size={16} className="text-text-secondary" />,
};

const TYPE_LABELS: Record<NotificationType, string> = {
  flag_change: 'Flag Change',
  alert: 'Alert',
  announcement: 'Announcement',
  system: 'System',
};

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [unreadCount, setUnreadCount] = useState(0);
  const [notificationToDelete, setNotificationToDelete] = useState<Notification | null>(null);
  const [markingAllRead, setMarkingAllRead] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [notifResponse, countResponse] = await Promise.all([
        notificationsApi.list(),
        notificationsApi.getUnreadCount(),
      ]);
      setNotifications(notifResponse.notifications || []);
      setUnreadCount(countResponse.count || 0);
      setError(null);
    } catch (err) {
      setError('Failed to load notifications');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleMarkAsRead = async (notification: Notification) => {
    if (notification.read_at) return;
    
    try {
      await notificationsApi.markAsRead(notification.id);
      setNotifications(prev => 
        prev.map(n => n.id === notification.id ? { ...n, read_at: new Date().toISOString() } : n)
      );
      setUnreadCount(prev => Math.max(0, prev - 1));
    } catch (err) {
      console.error('Failed to mark as read:', err);
    }
  };

  const handleMarkAllAsRead = async () => {
    if (unreadCount === 0) return;
    
    try {
      setMarkingAllRead(true);
      await notificationsApi.markAllAsRead();
      setNotifications(prev => 
        prev.map(n => ({ ...n, read_at: n.read_at || new Date().toISOString() }))
      );
      setUnreadCount(0);
    } catch (err) {
      console.error('Failed to mark all as read:', err);
    } finally {
      setMarkingAllRead(false);
    }
  };

  const handleDelete = async () => {
    if (!notificationToDelete) return;
    
    try {
      await notificationsApi.delete(notificationToDelete.id);
      setNotifications(prev => prev.filter(n => n.id !== notificationToDelete.id));
      if (!notificationToDelete.read_at) {
        setUnreadCount(prev => Math.max(0, prev - 1));
      }
      setNotificationToDelete(null);
    } catch (err) {
      console.error('Failed to delete notification:', err);
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          <Link
            to="/"
            className="p-2 hover:bg-bg-secondary rounded-lg transition-colors"
          >
            <ChevronLeft size={20} className="text-text-secondary" />
          </Link>
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
              <Bell className="text-accent-purple" />
              Notifications
              {unreadCount > 0 && (
                <span className="text-sm bg-accent-red text-white px-2 py-0.5 rounded-full">
                  {unreadCount} new
                </span>
              )}
            </h1>
            <p className="text-text-secondary mt-1">
              Stay updated on flag changes and system events
            </p>
          </div>
          {unreadCount > 0 && (
            <button
              onClick={handleMarkAllAsRead}
              disabled={markingAllRead}
              className="flex items-center gap-2 px-4 py-2 bg-bg-secondary hover:bg-bg-tertiary rounded-lg transition-colors disabled:opacity-50"
            >
              {markingAllRead ? (
                <Loader2 size={16} className="animate-spin" />
              ) : (
                <CheckCheck size={16} />
              )}
              <span>Mark all read</span>
            </button>
          )}
        </div>

        {/* Error State */}
        {error && (
          <div className="bg-accent-red/10 border border-accent-red/20 rounded-lg p-4 mb-6 flex items-center gap-3">
            <AlertCircle className="text-accent-red" size={20} />
            <span className="text-accent-red">{error}</span>
          </div>
        )}

        {/* Loading State */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="animate-spin text-accent-purple" size={32} />
          </div>
        ) : notifications.length === 0 ? (
          /* Empty State */
          <div className="text-center py-12 bg-bg-secondary rounded-lg">
            <Bell size={48} className="mx-auto text-text-tertiary mb-4" />
            <h3 className="text-lg font-medium text-text-primary">No notifications</h3>
            <p className="text-text-secondary mt-1">
              You're all caught up! Check back later for updates.
            </p>
          </div>
        ) : (
          /* Notifications List */
          <div className="space-y-2">
            {notifications.map(notification => (
              <div
                key={notification.id}
                className={`bg-bg-secondary rounded-lg p-4 hover:bg-bg-tertiary transition-colors ${
                  !notification.read_at ? 'border-l-4 border-accent-purple' : ''
                }`}
              >
                <div className="flex items-start gap-4">
                  <div className="p-2 bg-bg-tertiary rounded-lg">
                    {TYPE_ICONS[notification.type]}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs text-text-tertiary">
                        {TYPE_LABELS[notification.type]}
                      </span>
                      <span className="text-text-tertiary">•</span>
                      <span className="text-xs text-text-tertiary">
                        {formatDate(notification.created_at)}
                      </span>
                      {!notification.read_at && (
                        <span className="w-2 h-2 bg-accent-purple rounded-full" />
                      )}
                    </div>
                    <h3 className="font-medium text-text-primary truncate">
                      {notification.title}
                    </h3>
                    <p className="text-sm text-text-secondary mt-1 line-clamp-2">
                      {notification.message}
                    </p>
                    {notification.link && (
                      <a
                        href={notification.link}
                        className="inline-flex items-center gap-1 text-sm text-accent-blue hover:underline mt-2"
                      >
                        View details
                        <ExternalLink size={12} />
                      </a>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    {!notification.read_at && (
                      <button
                        onClick={() => handleMarkAsRead(notification)}
                        className="p-2 hover:bg-bg-primary rounded-lg transition-colors text-text-secondary hover:text-accent-green"
                        title="Mark as read"
                      >
                        <Check size={16} />
                      </button>
                    )}
                    <button
                      onClick={() => setNotificationToDelete(notification)}
                      className="p-2 hover:bg-bg-primary rounded-lg transition-colors text-text-secondary hover:text-accent-red"
                      title="Delete"
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Delete Confirmation Modal */}
      <ConfirmDeleteModal
        isOpen={!!notificationToDelete}
        onClose={() => setNotificationToDelete(null)}
        onConfirm={handleDelete}
        title="Delete Notification"
        message="Are you sure you want to delete this notification? This action cannot be undone."
      />
    </div>
  );
}
