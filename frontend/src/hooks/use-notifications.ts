import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import { useAuthStore } from '@/store/auth-store';
import type { Notification } from '@/types/api';

/** Fetch recent notifications for the authenticated user. */
export function useNotifications(limit = 20) {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.notifications.list(userId!),
    queryFn: () => apiClient.get<Notification[]>('/api/notifications', { limit: String(limit) }),
    staleTime: staleTimes.notifications,
    enabled: !!userId,
  });
}

/** Fetch unread notification count. */
export function useUnreadCount() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.notifications.unreadCount(userId!),
    queryFn: () => apiClient.get<{ count: number }>('/api/notifications/unread-count'),
    staleTime: staleTimes.notifications,
    refetchInterval: 30_000, // Poll every 30 seconds
    enabled: !!userId,
  });
}

/** Mark a notification as read. */
export function useMarkNotificationRead() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (id: string) => apiClient.put<{ message: string }>(`/api/notifications/${id}/read`),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.notifications.all(userId) });
    },
  });
}
