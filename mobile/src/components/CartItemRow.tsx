import React from 'react'
import { View, StyleSheet, TouchableOpacity } from 'react-native'
import { Text } from 'react-native-paper'
import { useRouter } from 'expo-router'
import { CartItem } from '../types'

interface CartItemRowProps {
  item: CartItem
  productName?: string
  productSlug?: string
  onUpdate: (quantity: number) => void
  onRemove: () => void
}

function formatPrice(price: number): string {
  return '₽ ' + price.toLocaleString('ru-RU')
}

export default function CartItemRow({ item, productName, productSlug, onUpdate, onRemove }: CartItemRowProps) {
  const router = useRouter()

  const handleNamePress = () => {
    if (productSlug) {
      router.push(`/product/${productSlug}` as never)
    }
  }

  return (
    <View style={styles.row}>
      <View style={styles.info}>
        <TouchableOpacity onPress={handleNamePress} disabled={!productSlug}>
          <Text numberOfLines={2} style={[styles.name, productSlug && styles.nameLink]}>
            {productName ?? `Variant #${item.variant_id.slice(-8)}`}
          </Text>
        </TouchableOpacity>
        <Text style={styles.price}>
          {formatPrice(item.price_snapshot)} × {item.quantity} ={' '}
          {formatPrice(item.price_snapshot * item.quantity)}
        </Text>
      </View>
      <View style={styles.controls}>
        <TouchableOpacity
          style={styles.qtyBtn}
          onPress={() => item.quantity > 1 && onUpdate(item.quantity - 1)}
          disabled={item.quantity <= 1}
        >
          <Text style={[styles.qtyBtnText, item.quantity <= 1 && styles.disabled]}>−</Text>
        </TouchableOpacity>
        <Text style={styles.qty}>{item.quantity}</Text>
        <TouchableOpacity
          style={styles.qtyBtn}
          onPress={() => onUpdate(item.quantity + 1)}
        >
          <Text style={styles.qtyBtnText}>+</Text>
        </TouchableOpacity>
        <TouchableOpacity style={styles.deleteBtn} onPress={onRemove}>
          <Text style={styles.deleteBtnText}>🗑</Text>
        </TouchableOpacity>
      </View>
    </View>
  )
}

const styles = StyleSheet.create({
  row: {
    backgroundColor: '#FFFFFF',
    borderRadius: 10,
    padding: 12,
    marginBottom: 10,
    flexDirection: 'column',
    gap: 8,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.06,
    shadowRadius: 2,
  },
  info: {
    gap: 2,
  },
  name: {
    fontSize: 14,
    fontWeight: '600',
    color: '#111827',
  },
  nameLink: {
    color: '#1E90FF',
    textDecorationLine: 'underline',
  },
  sku: {
    fontSize: 12,
    color: '#9CA3AF',
  },
  price: {
    fontSize: 13,
    color: '#374151',
    marginTop: 2,
  },
  controls: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  qtyBtn: {
    width: 32,
    height: 32,
    borderRadius: 6,
    backgroundColor: '#F3F4F6',
    alignItems: 'center',
    justifyContent: 'center',
  },
  qtyBtnText: {
    fontSize: 18,
    color: '#1E90FF',
    fontWeight: '600',
    lineHeight: 22,
  },
  disabled: {
    color: '#D1D5DB',
  },
  qty: {
    fontSize: 15,
    fontWeight: '700',
    color: '#111827',
    minWidth: 24,
    textAlign: 'center',
  },
  deleteBtn: {
    marginLeft: 'auto',
    width: 32,
    height: 32,
    borderRadius: 6,
    backgroundColor: '#FEF2F2',
    alignItems: 'center',
    justifyContent: 'center',
  },
  deleteBtnText: {
    fontSize: 16,
  },
})
