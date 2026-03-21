import React from 'react'
import { View, TouchableOpacity, Image, StyleSheet, Dimensions } from 'react-native'
import { Text } from 'react-native-paper'
import { useRouter } from 'expo-router'
import { Product } from '../types'
import StarRating from './StarRating'

interface ProductCardProps {
  product: Product
}

const CARD_WIDTH = Dimensions.get('window').width / 2 - 24

function formatPrice(price: number): string {
  return '₽ ' + price.toLocaleString('ru-RU')
}

function getMinPrice(product: Product): number | null {
  if (!product.variants || product.variants.length === 0) return null
  return Math.min(...product.variants.map((v) => v.price))
}

export default function ProductCard({ product }: ProductCardProps) {
  const router = useRouter()
  const minPrice = getMinPrice(product)
  const imageUrl = product.primary_image_url
    ?? product.images?.find((img) => img.sort_order === 0)?.url
    ?? product.images?.[0]?.url

  return (
    <TouchableOpacity
      style={styles.card}
      onPress={() => router.push(`/product/${product.slug}` as never)}
      activeOpacity={0.85}
    >
      {imageUrl ? (
        <Image
          source={{ uri: imageUrl }}
          style={styles.image}
          resizeMode="cover"
        />
      ) : (
        <View style={[styles.image, styles.imagePlaceholder]} />
      )}
      <View style={styles.info}>
        <Text numberOfLines={2} style={styles.name}>
          {product.name}
        </Text>
        {minPrice !== null && (
          <Text style={styles.price}>{formatPrice(minPrice)}</Text>
        )}
        {product.rating > 0 && (
          <View style={styles.ratingRow}>
            <StarRating rating={product.rating} size={14} />
            {product.review_count > 0 && (
              <Text style={styles.reviewCount}>({product.review_count})</Text>
            )}
          </View>
        )}
      </View>
    </TouchableOpacity>
  )
}

const styles = StyleSheet.create({
  card: {
    width: CARD_WIDTH,
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    marginBottom: 12,
    overflow: 'hidden',
    elevation: 2,
    shadowColor: '#1E90FF',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.10,
    shadowRadius: 8,
  },
  image: {
    width: '100%',
    height: 150,
  },
  imagePlaceholder: {
    backgroundColor: '#E5E7EB',
  },
  info: {
    padding: 8,
    gap: 4,
  },
  name: {
    fontSize: 13,
    fontWeight: '600',
    color: '#111827',
    lineHeight: 18,
  },
  price: {
    fontSize: 14,
    fontWeight: '700',
    color: '#1E90FF',
    marginTop: 2,
  },
  ratingRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    marginTop: 2,
  },
  reviewCount: {
    fontSize: 11,
    color: '#9CA3AF',
  },
})
