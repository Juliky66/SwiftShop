import React, { useState } from 'react'
import {
  View,
  ScrollView,
  StyleSheet,
  TouchableOpacity,
} from 'react-native'
import { Text, Button, Divider, Chip, Snackbar } from 'react-native-paper'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useLocalSearchParams, useRouter } from 'expo-router'
import { SafeAreaView } from 'react-native-safe-area-context'
import { MaterialCommunityIcons } from '@expo/vector-icons'
import { getOrder, cancelOrder } from '../../src/api/orders'
import { initiatePayment } from '../../src/api/reviews'
import { OrderItem } from '../../src/types'
import LoadingSpinner from '../../src/components/LoadingSpinner'

function formatPrice(price: number): string {
  return '₽ ' + price.toLocaleString('ru-RU')
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'pending': return '#F59E0B'
    case 'paid': return '#10B981'
    case 'cancelled': return '#EF4444'
    case 'processing': return '#3B82F6'
    default: return '#6B7280'
  }
}

function getStatusLabel(status: string): string {
  switch (status) {
    case 'pending': return 'Pending Payment'
    case 'paid': return 'Paid'
    case 'cancelled': return 'Cancelled'
    case 'processing': return 'Processing'
    default: return status
  }
}

export default function OrderDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>()
  const router = useRouter()
  const queryClient = useQueryClient()

  const [snackMsg, setSnackMsg] = useState('')
  const [snackVisible, setSnackVisible] = useState(false)

  const showSnack = (msg: string) => {
    setSnackMsg(msg)
    setSnackVisible(true)
  }

  const orderQuery = useQuery({
    queryKey: ['order', id],
    queryFn: () => getOrder(id),
    enabled: !!id,
  })

  const cancelMutation = useMutation({
    mutationFn: () => cancelOrder(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['order', id] })
      queryClient.invalidateQueries({ queryKey: ['orders'] })
      showSnack('Order cancelled.')
    },
    onError: () => showSnack('Could not cancel order. Please try again.'),
  })

  const payMutation = useMutation({
    mutationFn: () => initiatePayment(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['order', id] })
      queryClient.invalidateQueries({ queryKey: ['orders'] })
      showSnack('Payment initiated! Your order is being processed.')
    },
    onError: () => showSnack('Could not initiate payment. Please try again.'),
  })

  if (orderQuery.isLoading) {
    return <LoadingSpinner message="Loading order..." />
  }

  const order = orderQuery.data

  if (!order) {
    return (
      <SafeAreaView style={styles.container} edges={['top']}>
        <View style={styles.backHeader}>
          <TouchableOpacity onPress={() => router.back()} style={styles.backButton}>
            <MaterialCommunityIcons name="arrow-left" size={24} color="#FFFFFF" />
          </TouchableOpacity>
          <Text style={styles.backHeaderTitle}>Order Details</Text>
        </View>
        <View style={styles.errorState}>
          <Text style={styles.errorText}>Order not found.</Text>
        </View>
      </SafeAreaView>
    )
  }

  const statusColor = getStatusColor(order.status)
  const items: OrderItem[] = order.items ?? []

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.backHeader}>
        <TouchableOpacity onPress={() => router.back()} style={styles.backButton}>
          <MaterialCommunityIcons name="arrow-left" size={24} color="#FFFFFF" />
        </TouchableOpacity>
        <Text style={styles.backHeaderTitle}>Order Details</Text>
      </View>

      <Snackbar
        visible={snackVisible}
        onDismiss={() => setSnackVisible(false)}
        duration={3000}
        style={{ backgroundColor: '#1F2937' }}
        action={{ label: 'OK', onPress: () => setSnackVisible(false) }}
      >
        {snackMsg}
      </Snackbar>

      <ScrollView style={styles.scroll} showsVerticalScrollIndicator={false}>
        <View style={styles.content}>
          {/* Order ID + Status */}
          <View style={styles.statusRow}>
            <View>
              <Text style={styles.orderIdLabel}>Order ID</Text>
              <Text style={styles.orderId}>#{order.id.slice(-12).toUpperCase()}</Text>
              <Text style={styles.orderDate}>{formatDate(order.created_at)}</Text>
            </View>
            <Chip
              style={[
                styles.statusChip,
                { backgroundColor: statusColor + '22' },
              ]}
              textStyle={[styles.statusText, { color: statusColor }]}
            >
              {getStatusLabel(order.status)}
            </Chip>
          </View>

          <Divider style={styles.divider} />

          {/* Items */}
          <Text style={styles.sectionTitle}>Items ({items.length})</Text>
          {items.map((item, idx) => (
            <View key={`${item.variant_id}-${idx}`} style={styles.itemRow}>
              <View style={styles.itemLeft}>
                <Text style={styles.itemName} numberOfLines={2}>
                  {item.product_name}
                </Text>
                <Text style={styles.itemSku}>SKU: {item.sku}</Text>
              </View>
              <View style={styles.itemRight}>
                <Text style={styles.itemQty}>×{item.quantity}</Text>
                <Text style={styles.itemPrice}>{formatPrice(item.unit_price * item.quantity)}</Text>
              </View>
            </View>
          ))}

          <Divider style={styles.divider} />

          {/* Address */}
          <View style={styles.addressSection}>
            <Text style={styles.sectionTitle}>Delivery Address</Text>
            <View style={styles.addressBox}>
              <MaterialCommunityIcons name="map-marker-outline" size={18} color="#1E90FF" />
              <Text style={styles.addressText}>{(() => {
                try {
                  const a = JSON.parse(order.shipping_address)
                  return `${a.full_name}\n${a.street}, ${a.city} ${a.postal_code}\n${a.phone}`
                } catch {
                  return order.shipping_address
                }
              })()}</Text>
            </View>
          </View>

          <Divider style={styles.divider} />

          {/* Total */}
          <View style={styles.totalRow}>
            <Text style={styles.totalLabel}>Order Total</Text>
            <Text style={styles.totalAmount}>{formatPrice(order.total_amount)}</Text>
          </View>

          {/* Actions */}
          <View style={styles.actionsSection}>
            {order.status === 'pending' && (
              <Button
                mode="contained"
                onPress={() => payMutation.mutate()}
                loading={payMutation.isPending}
                disabled={payMutation.isPending}
                style={styles.actionButton}
                buttonColor="#10B981"
                contentStyle={styles.actionButtonContent}
                labelStyle={styles.actionButtonLabel}
                icon="credit-card-outline"
              >
                Pay Now
              </Button>
            )}
            {order.status === 'processing' && (
              <Button
                mode="outlined"
                onPress={() => cancelMutation.mutate()}
                loading={cancelMutation.isPending}
                disabled={cancelMutation.isPending}
                style={styles.cancelButton}
                textColor="#EF4444"
                contentStyle={styles.actionButtonContent}
                icon="close-circle-outline"
              >
                Cancel Order
              </Button>
            )}
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F8FAFF',
  },
  backHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1E90FF',
    paddingHorizontal: 12,
    paddingVertical: 10,
    gap: 10,
  },
  backButton: {
    padding: 4,
  },
  backHeaderTitle: {
    fontSize: 17,
    fontWeight: '700',
    color: '#FFFFFF',
  },
  scroll: {
    flex: 1,
  },
  content: {
    padding: 16,
    gap: 12,
  },
  statusRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
  },
  orderIdLabel: {
    fontSize: 11,
    color: '#9CA3AF',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  orderId: {
    fontSize: 16,
    fontWeight: '700',
    color: '#111827',
    fontFamily: 'monospace',
    marginTop: 2,
  },
  orderDate: {
    fontSize: 12,
    color: '#9CA3AF',
    marginTop: 2,
  },
  statusChip: {
    borderRadius: 8,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '700',
  },
  divider: {
    backgroundColor: '#E5E7EB',
    marginVertical: 4,
  },
  sectionTitle: {
    fontSize: 15,
    fontWeight: '700',
    color: '#111827',
    marginBottom: 8,
  },
  itemRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    backgroundColor: '#FFFFFF',
    borderRadius: 10,
    padding: 12,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#F3F4F6',
  },
  itemLeft: {
    flex: 1,
    marginRight: 12,
    gap: 3,
  },
  itemName: {
    fontSize: 14,
    fontWeight: '600',
    color: '#111827',
    lineHeight: 20,
  },
  itemSku: {
    fontSize: 11,
    color: '#9CA3AF',
    fontFamily: 'monospace',
  },
  itemRight: {
    alignItems: 'flex-end',
    gap: 3,
  },
  itemQty: {
    fontSize: 13,
    color: '#6B7280',
  },
  itemPrice: {
    fontSize: 14,
    fontWeight: '700',
    color: '#374151',
  },
  addressSection: {
    gap: 8,
  },
  addressBox: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    backgroundColor: '#FFFFFF',
    borderRadius: 10,
    padding: 12,
    gap: 8,
    borderWidth: 1,
    borderColor: '#F3F4F6',
  },
  addressText: {
    flex: 1,
    fontSize: 14,
    color: '#374151',
    lineHeight: 22,
  },
  totalRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: '#FFFFFF',
    borderRadius: 10,
    padding: 14,
    borderWidth: 1,
    borderColor: '#E5E7EB',
  },
  totalLabel: {
    fontSize: 16,
    fontWeight: '500',
    color: '#374151',
  },
  totalAmount: {
    fontSize: 22,
    fontWeight: '800',
    color: '#1E90FF',
  },
  actionsSection: {
    gap: 10,
    paddingBottom: 24,
  },
  actionButton: {
    borderRadius: 12,
  },
  cancelButton: {
    borderRadius: 12,
    borderColor: '#EF4444',
  },
  actionButtonContent: {
    height: 50,
  },
  actionButtonLabel: {
    fontSize: 15,
    fontWeight: '700',
  },
  errorState: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
  },
  errorText: {
    fontSize: 16,
    color: '#6B7280',
  },
})
