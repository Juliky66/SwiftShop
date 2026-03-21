import React, { useState } from 'react'
import { View, FlatList, StyleSheet } from 'react-native'
import { Text, Button, Divider, Dialog, Portal, TextInput, Snackbar } from 'react-native-paper'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'expo-router'
import { SafeAreaView } from 'react-native-safe-area-context'
import {
  getCart,
  updateCartItem,
  removeCartItem,
  checkout,
} from '../../src/api/orders'
import { CartItem, DeliveryAddress } from '../../src/types'
import CartItemRow from '../../src/components/CartItemRow'
import { useCartEnrichment } from '../../src/store/cartEnrichment'
import LoadingSpinner from '../../src/components/LoadingSpinner'

function formatPrice(price: number): string {
  return '₽ ' + price.toLocaleString('ru-RU')
}

export default function CartScreen() {
  const queryClient = useQueryClient()
  const router = useRouter()
  const { items: cartEnrichment } = useCartEnrichment()
  const [checkoutVisible, setCheckoutVisible] = useState(false)
  const [snackMsg, setSnackMsg] = useState('')
  const [snackVisible, setSnackVisible] = useState(false)
  const [addrFullName, setAddrFullName] = useState('')
  const [addrPhone, setAddrPhone] = useState('')
  const [addrCity, setAddrCity] = useState('')
  const [addrStreet, setAddrStreet] = useState('')
  const [addrPostal, setAddrPostal] = useState('')
  const [addressError, setAddressError] = useState('')

  const cartQuery = useQuery({
    queryKey: ['cart'],
    queryFn: getCart,
  })

  const updateMutation = useMutation({
    mutationFn: ({ variantId, quantity }: { variantId: string; quantity: number }) =>
      updateCartItem(variantId, quantity),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['cart'] }),
  })

  const removeMutation = useMutation({
    mutationFn: (variantId: string) => removeCartItem(variantId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['cart'] }),
  })

  const checkoutMutation = useMutation({
    mutationFn: (addr: DeliveryAddress) => checkout(addr),
    onSuccess: (order) => {
      queryClient.invalidateQueries({ queryKey: ['cart'] })
      queryClient.invalidateQueries({ queryKey: ['orders'] })
      setCheckoutVisible(false)
      router.push(`/order/${order.id}` as never)
    },
    onError: () => {
      setSnackMsg('Checkout failed. Please try again.')
      setSnackVisible(true)
    },
  })

  const handleCheckout = () => {
    if (!addrFullName.trim() || !addrCity.trim() || !addrStreet.trim()) {
      setAddressError('Please fill in name, city, and street.')
      return
    }
    setAddressError('')
    checkoutMutation.mutate({
      full_name: addrFullName.trim(),
      phone: addrPhone.trim(),
      city: addrCity.trim(),
      street: addrStreet.trim(),
      postal_code: addrPostal.trim(),
    })
  }

  if (cartQuery.isLoading) {
    return <LoadingSpinner message="Loading cart..." />
  }

  const cartItems: CartItem[] = cartQuery.data?.items ?? []
  const total = cartItems.reduce((sum, item) => sum + item.price_snapshot * item.quantity, 0)

  if (cartItems.length === 0) {
    return (
      <SafeAreaView style={styles.container} edges={['top']}>
        <View style={styles.header}>
          <Text style={styles.headerTitle}>My Cart</Text>
        </View>
        <View style={styles.emptyState}>
          <Text style={styles.emptyIcon}>🛒</Text>
          <Text style={styles.emptyText}>Your cart is empty</Text>
          <Text style={styles.emptySubtext}>Add some products to get started</Text>
          <Button
            mode="contained"
            onPress={() => router.push('/(tabs)' as never)}
            style={styles.browseButton}
            buttonColor="#1E90FF"
          >
            Browse Products
          </Button>
        </View>
      </SafeAreaView>
    )
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>My Cart</Text>
        <Text style={styles.headerCount}>{cartItems.length} items</Text>
      </View>

      <FlatList
        data={cartItems}
        keyExtractor={(item) => item.variant_id}
        renderItem={({ item }) => (
          <CartItemRow
            item={item}
            productName={cartEnrichment[item.variant_id]?.name}
            productSlug={cartEnrichment[item.variant_id]?.slug}
            onUpdate={(qty) =>
              updateMutation.mutate({ variantId: item.variant_id, quantity: qty })
            }
            onRemove={() => removeMutation.mutate(item.variant_id)}
          />
        )}
        contentContainerStyle={styles.list}
        showsVerticalScrollIndicator={false}
      />

      <View style={styles.footer}>
        <Divider />
        <View style={styles.totalRow}>
          <Text style={styles.totalLabel}>Total</Text>
          <Text style={styles.totalAmount}>{formatPrice(total)}</Text>
        </View>
        <Button
          mode="contained"
          onPress={() => setCheckoutVisible(true)}
          style={styles.checkoutButton}
          buttonColor="#1E90FF"
          contentStyle={styles.checkoutContent}
          labelStyle={styles.checkoutLabel}
        >
          Checkout
        </Button>
      </View>

      <Portal>
        <Dialog
          visible={checkoutVisible}
          onDismiss={() => setCheckoutVisible(false)}
          style={styles.dialog}
        >
          <Dialog.Title>Delivery Address</Dialog.Title>
          <Dialog.Content>
            <TextInput label="Full Name *" value={addrFullName} onChangeText={setAddrFullName}
              mode="outlined" outlineColor="#E5E7EB" activeOutlineColor="#1E90FF"
              style={styles.addressInput} />
            <TextInput label="Phone" value={addrPhone} onChangeText={setAddrPhone}
              mode="outlined" outlineColor="#E5E7EB" activeOutlineColor="#1E90FF"
              style={styles.addressInput} keyboardType="phone-pad" />
            <TextInput label="City *" value={addrCity} onChangeText={setAddrCity}
              mode="outlined" outlineColor="#E5E7EB" activeOutlineColor="#1E90FF"
              style={styles.addressInput} />
            <TextInput label="Street & Building *" value={addrStreet} onChangeText={setAddrStreet}
              mode="outlined" outlineColor="#E5E7EB" activeOutlineColor="#1E90FF"
              style={styles.addressInput} />
            <TextInput label="Postal Code" value={addrPostal} onChangeText={setAddrPostal}
              mode="outlined" outlineColor="#E5E7EB" activeOutlineColor="#1E90FF"
              style={styles.addressInput} keyboardType="numeric" />
            {addressError ? (
              <Text style={styles.inputError}>{addressError}</Text>
            ) : null}
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={() => setCheckoutVisible(false)} textColor="#6B7280">
              Cancel
            </Button>
            <Button
              onPress={handleCheckout}
              loading={checkoutMutation.isPending}
              disabled={checkoutMutation.isPending}
              textColor="#1E90FF"
            >
              Place Order
            </Button>
          </Dialog.Actions>
        </Dialog>
      </Portal>

      <Snackbar
        visible={snackVisible}
        onDismiss={() => setSnackVisible(false)}
        duration={3000}
        style={{ backgroundColor: '#1F2937' }}
        action={{ label: 'OK', onPress: () => setSnackVisible(false) }}
      >
        {snackMsg}
      </Snackbar>
    </SafeAreaView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F8FAFF',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 14,
    backgroundColor: '#1E90FF',
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#FFFFFF',
  },
  headerCount: {
    fontSize: 14,
    color: '#DDD6FE',
  },
  emptyState: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
    paddingHorizontal: 32,
  },
  emptyIcon: {
    fontSize: 56,
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 20,
    fontWeight: '700',
    color: '#374151',
  },
  emptySubtext: {
    fontSize: 14,
    color: '#9CA3AF',
    marginBottom: 8,
  },
  browseButton: {
    borderRadius: 10,
    marginTop: 8,
  },
  list: {
    padding: 12,
    paddingBottom: 0,
  },
  footer: {
    backgroundColor: '#FFFFFF',
    paddingHorizontal: 16,
    paddingTop: 12,
    paddingBottom: 16,
    elevation: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: -3 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
  },
  totalRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 12,
  },
  totalLabel: {
    fontSize: 16,
    color: '#374151',
    fontWeight: '500',
  },
  totalAmount: {
    fontSize: 22,
    fontWeight: '800',
    color: '#1E90FF',
  },
  checkoutButton: {
    borderRadius: 10,
  },
  checkoutContent: {
    height: 50,
  },
  checkoutLabel: {
    fontSize: 16,
    fontWeight: '700',
  },
  dialog: {
    borderRadius: 16,
  },
  dialogSubtext: {
    fontSize: 13,
    color: '#6B7280',
    marginBottom: 12,
  },
  addressInput: {
    backgroundColor: '#FFFFFF',
    marginTop: 4,
  },
  inputError: {
    color: '#EF4444',
    fontSize: 12,
    marginTop: 4,
  },
})
