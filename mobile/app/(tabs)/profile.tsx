import React from 'react'
import { View, FlatList, StyleSheet, TouchableOpacity } from 'react-native'
import { Text, Button, Chip, Divider, Surface } from 'react-native-paper'
import { useQuery } from '@tanstack/react-query'
import { useRouter } from 'expo-router'
import { SafeAreaView } from 'react-native-safe-area-context'
import { getOrders } from '../../src/api/orders'
import { Order } from '../../src/types'
import { useAuthStore } from '../../src/store/auth'
import LoadingSpinner from '../../src/components/LoadingSpinner'

function getStatusColor(status: string): string {
  switch (status) {
    case 'pending': return '#F59E0B'
    case 'paid': return '#10B981'
    case 'cancelled': return '#EF4444'
    case 'processing': return '#3B82F6'
    default: return '#6B7280'
  }
}

function formatPrice(price: number): string {
  return '₽ ' + price.toLocaleString('ru-RU')
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: '2-digit',
  })
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .slice(0, 2)
    .map((n) => n[0]?.toUpperCase() ?? '')
    .join('')
}

export default function ProfileScreen() {
  const { user, logout } = useAuthStore()
  const router = useRouter()

  const ordersQuery = useQuery({
    queryKey: ['orders'],
    queryFn: () => getOrders({ limit: 20 }),
  })

  const orders: Order[] = ordersQuery.data?.items ?? []

  const renderOrder = ({ item }: { item: Order }) => (
    <TouchableOpacity
      style={styles.orderRow}
      onPress={() => router.push(`/order/${item.id}` as never)}
      activeOpacity={0.7}
    >
      <View style={styles.orderLeft}>
        <Text style={styles.orderId}>#{item.id.slice(-8).toUpperCase()}</Text>
        <Text style={styles.orderDate}>{formatDate(item.created_at)}</Text>
      </View>
      <View style={styles.orderRight}>
        <Chip
          style={[styles.statusChip, { backgroundColor: getStatusColor(item.status) + '22' }]}
          textStyle={[styles.statusText, { color: getStatusColor(item.status) }]}
          compact
        >
          {item.status}
        </Chip>
        <Text style={styles.orderTotal}>{formatPrice(item.total_amount)}</Text>
      </View>
    </TouchableOpacity>
  )

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Profile</Text>
      </View>

      {/* User Info Card */}
      {user && (
        <Surface style={styles.userCard} elevation={1}>
          <View style={styles.avatarCircle}>
            <Text style={styles.avatarText}>{getInitials(user.full_name)}</Text>
          </View>
          <View style={styles.userInfo}>
            <Text style={styles.userName}>{user.full_name}</Text>
            <Text style={styles.userEmail}>{user.email}</Text>
            <View style={styles.roleRow}>
              <Chip
                style={styles.roleChip}
                textStyle={styles.roleChipText}
                compact
              >
                {user.role}
              </Chip>
              {user.phone && (
                <Text style={styles.userPhone}>{user.phone}</Text>
              )}
            </View>
          </View>
        </Surface>
      )}

      <Divider style={styles.divider} />

      {/* Orders section */}
      <View style={styles.ordersSection}>
        <Text style={styles.sectionTitle}>My Orders</Text>
        {ordersQuery.isLoading ? (
          <LoadingSpinner message="Loading orders..." />
        ) : orders.length === 0 ? (
          <View style={styles.noOrders}>
            <Text style={styles.noOrdersIcon}>📦</Text>
            <Text style={styles.noOrdersText}>No orders yet</Text>
          </View>
        ) : (
          <FlatList
            data={orders}
            renderItem={renderOrder}
            keyExtractor={(item) => item.id}
            scrollEnabled={false}
            ItemSeparatorComponent={() => <View style={styles.separator} />}
          />
        )}
      </View>

      <View style={styles.logoutContainer}>
        <Button
          mode="outlined"
          onPress={logout}
          textColor="#EF4444"
          style={styles.logoutButton}
          icon="logout"
        >
          Sign Out
        </Button>
      </View>
    </SafeAreaView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F8FAFF',
  },
  header: {
    paddingHorizontal: 16,
    paddingVertical: 14,
    backgroundColor: '#1E90FF',
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#FFFFFF',
  },
  userCard: {
    margin: 16,
    padding: 16,
    borderRadius: 16,
    backgroundColor: '#FFFFFF',
    flexDirection: 'row',
    alignItems: 'center',
    gap: 14,
  },
  avatarCircle: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: '#1E90FF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  avatarText: {
    color: '#FFFFFF',
    fontSize: 22,
    fontWeight: '700',
  },
  userInfo: {
    flex: 1,
    gap: 3,
  },
  userName: {
    fontSize: 17,
    fontWeight: '700',
    color: '#111827',
  },
  userEmail: {
    fontSize: 13,
    color: '#6B7280',
  },
  roleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginTop: 4,
  },
  roleChip: {
    backgroundColor: '#EDE9FE',
    height: 24,
  },
  roleChipText: {
    color: '#1E90FF',
    fontSize: 11,
    fontWeight: '600',
  },
  userPhone: {
    fontSize: 12,
    color: '#9CA3AF',
  },
  divider: {
    marginHorizontal: 16,
    backgroundColor: '#E5E7EB',
  },
  ordersSection: {
    flex: 1,
    paddingHorizontal: 16,
    paddingTop: 16,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#111827',
    marginBottom: 12,
  },
  orderRow: {
    backgroundColor: '#FFFFFF',
    borderRadius: 10,
    padding: 14,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  orderLeft: {
    gap: 3,
  },
  orderId: {
    fontSize: 13,
    fontWeight: '700',
    color: '#111827',
    fontFamily: 'monospace',
  },
  orderDate: {
    fontSize: 12,
    color: '#9CA3AF',
  },
  orderRight: {
    alignItems: 'flex-end',
    gap: 4,
  },
  statusChip: {
    height: 22,
  },
  statusText: {
    fontSize: 11,
    fontWeight: '600',
  },
  orderTotal: {
    fontSize: 14,
    fontWeight: '700',
    color: '#374151',
  },
  separator: {
    height: 8,
  },
  noOrders: {
    alignItems: 'center',
    paddingVertical: 32,
    gap: 8,
  },
  noOrdersIcon: {
    fontSize: 40,
  },
  noOrdersText: {
    fontSize: 16,
    color: '#9CA3AF',
  },
  logoutContainer: {
    padding: 16,
    paddingBottom: 24,
  },
  logoutButton: {
    borderColor: '#EF4444',
    borderRadius: 10,
  },
})
