import React, { useState } from 'react'
import {
  View,
  Image,
  ScrollView,
  StyleSheet,
  TouchableOpacity,
} from 'react-native'
import { Text, Button, Divider, Snackbar, Chip } from 'react-native-paper'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useLocalSearchParams, useRouter } from 'expo-router'
import { SafeAreaView } from 'react-native-safe-area-context'
import { MaterialCommunityIcons } from '@expo/vector-icons'
import { getProduct } from '../../src/api/catalog'
import { addToCart } from '../../src/api/orders'
import { useCartEnrichment } from '../../src/store/cartEnrichment'
import { getReviews } from '../../src/api/reviews'
import { ProductVariant, Review } from '../../src/types'
import StarRating from '../../src/components/StarRating'
import LoadingSpinner from '../../src/components/LoadingSpinner'

function formatPrice(price: number): string {
  return '₽ ' + price.toLocaleString('ru-RU')
}

function formatAttributes(attrs: Record<string, string>): string {
  return Object.entries(attrs)
    .map(([k, v]) => `${k}: ${v}`)
    .join(', ')
}

export default function ProductDetailScreen() {
  const { slug } = useLocalSearchParams<{ slug: string }>()
  const router = useRouter()
  const queryClient = useQueryClient()
  const { setItem: saveCartItem } = useCartEnrichment()
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null)
  const [snackbarVisible, setSnackbarVisible] = useState(false)
  const [snackbarMsg, setSnackbarMsg] = useState('')
  const [addedToCart, setAddedToCart] = useState(false)

  const productQuery = useQuery({
    queryKey: ['product', slug],
    queryFn: () => getProduct(slug),
    enabled: !!slug,
  })

  const product = productQuery.data

  const reviewsQuery = useQuery({
    queryKey: ['reviews', product?.id],
    queryFn: () => getReviews(product!.id, { limit: 20 }),
    enabled: !!product?.id,
  })

  const addToCartMutation = useMutation({
    mutationFn: () => addToCart(activeVariant!.id, 1),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cart'] })
      if (activeVariant && product) {
        saveCartItem(activeVariant.id, { name: product.name, slug: product.slug })
      }
      setSnackbarMsg('Added to cart!')
      setSnackbarVisible(true)
      setAddedToCart(true)
      setTimeout(() => setAddedToCart(false), 3000)
    },
    onError: (err: unknown) => {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setSnackbarMsg(msg ? `Error: ${msg}` : 'Failed to add to cart. Please sign in first.')
      setSnackbarVisible(true)
    },
  })

  if (productQuery.isLoading) {
    return <LoadingSpinner message="Loading product..." />
  }

  if (!product) {
    return (
      <SafeAreaView style={styles.container} edges={['top']}>
        <View style={styles.backHeader}>
          <TouchableOpacity onPress={() => router.back()} style={styles.backButton}>
            <MaterialCommunityIcons name="arrow-left" size={24} color="#111827" />
          </TouchableOpacity>
        </View>
        <View style={styles.errorState}>
          <Text style={styles.errorText}>Product not found.</Text>
        </View>
      </SafeAreaView>
    )
  }

  const firstImage = product.images?.find((img) => img.sort_order === 0)
    ?? product.images?.find((img) => img.sort_order === 1)
    ?? product.images?.[0]
  const reviews: Review[] = reviewsQuery.data?.items ?? []
  const activeVariant = selectedVariant ?? product.variants?.[0]

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      {/* Back header */}
      <View style={styles.backHeader}>
        <TouchableOpacity onPress={() => router.back()} style={styles.backButton}>
          <MaterialCommunityIcons name="arrow-left" size={24} color="#FFFFFF" />
        </TouchableOpacity>
        <Text style={styles.backHeaderTitle} numberOfLines={1}>
          {product.name}
        </Text>
      </View>

      <ScrollView showsVerticalScrollIndicator={false} style={styles.scroll}>
        {/* Product image */}
        {firstImage?.url ? (
          <Image
            source={{ uri: firstImage.url }}
            style={styles.productImage}
            resizeMode="cover"
          />
        ) : (
          <View style={[styles.productImage, styles.imagePlaceholder]} />
        )}

        <View style={styles.content}>
          {/* Name & rating */}
          <Text style={styles.productName}>{product.name}</Text>

          {product.rating > 0 && (
            <View style={styles.ratingRow}>
              <StarRating rating={product.rating} size={18} />
              <Text style={styles.ratingText}>
                {product.rating.toFixed(1)} ({product.review_count} reviews)
              </Text>
            </View>
          )}

          {/* Description */}
          {product.description ? (
            <Text style={styles.description}>{product.description}</Text>
          ) : null}

          {/* Variants */}
          {product.variants && product.variants.length > 0 && (
            <View style={styles.variantsSection}>
              <Text style={styles.sectionTitle}>Options</Text>
              <ScrollView
                horizontal
                showsHorizontalScrollIndicator={false}
                contentContainerStyle={styles.variantsScroll}
              >
                {product.variants.map((variant) => {
                  const isActive =
                    activeVariant?.id === variant.id
                  return (
                    <TouchableOpacity
                      key={variant.id}
                      style={[
                        styles.variantChip,
                        isActive && styles.variantChipActive,
                        variant.stock === 0 && styles.variantChipOos,
                      ]}
                      onPress={() => setSelectedVariant(variant)}
                      disabled={variant.stock === 0}
                    >
                      <Text
                        style={[
                          styles.variantPrice,
                          isActive && styles.variantPriceActive,
                        ]}
                      >
                        {formatPrice(variant.price)}
                      </Text>
                      {Object.keys(variant.attributes).length > 0 && (
                        <Text
                          style={[
                            styles.variantAttrs,
                            isActive && styles.variantAttrsActive,
                          ]}
                          numberOfLines={1}
                        >
                          {formatAttributes(variant.attributes)}
                        </Text>
                      )}
                      {variant.stock === 0 && (
                        <Text style={styles.oosLabel}>Out of stock</Text>
                      )}
                    </TouchableOpacity>
                  )
                })}
              </ScrollView>
            </View>
          )}

          {/* Add to cart */}
          <Button
            mode="contained"
            onPress={() => addToCartMutation.mutate()}
            loading={addToCartMutation.isPending}
            disabled={addToCartMutation.isPending || addedToCart || (activeVariant?.stock ?? 0) === 0}
            style={styles.addButton}
            buttonColor={addedToCart ? '#059669' : '#1E90FF'}
            contentStyle={styles.addButtonContent}
            labelStyle={styles.addButtonLabel}
            icon={addedToCart ? 'check' : 'cart-plus'}
          >
            {(activeVariant?.stock ?? 0) === 0 ? 'Out of Stock' : addedToCart ? 'Added!' : 'Add to Cart'}
          </Button>

          <Divider style={styles.divider} />

          {/* Reviews */}
          <View style={styles.reviewsSection}>
            <Text style={styles.sectionTitle}>
              Reviews ({reviewsQuery.data?.total ?? 0})
            </Text>
            {reviewsQuery.isLoading && (
              <LoadingSpinner message="Loading reviews..." />
            )}
            {reviews.length === 0 && !reviewsQuery.isLoading && (
              <Text style={styles.noReviews}>No reviews yet. Be the first!</Text>
            )}
            {reviews.map((review) => (
              <View key={review.id} style={styles.reviewCard}>
                <View style={styles.reviewHeader}>
                  <Text style={styles.reviewUser}>
                    User #{review.user_id.slice(-6)}
                  </Text>
                  <StarRating rating={review.rating} size={14} />
                </View>
                {review.title && (
                  <Text style={styles.reviewTitle}>{review.title}</Text>
                )}
                {review.body && (
                  <Text style={styles.reviewBody}>{review.body}</Text>
                )}
              </View>
            ))}
          </View>
        </View>
      </ScrollView>

      <Snackbar
        visible={snackbarVisible}
        onDismiss={() => setSnackbarVisible(false)}
        duration={2500}
        style={styles.snackbar}
        action={{ label: 'OK', onPress: () => setSnackbarVisible(false) }}
      >
        {snackbarMsg}
      </Snackbar>
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
    flex: 1,
    fontSize: 16,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  scroll: {
    flex: 1,
  },
  productImage: {
    width: '100%',
    height: 250,
  },
  imagePlaceholder: {
    backgroundColor: '#E5E7EB',
  },
  content: {
    padding: 16,
    gap: 12,
  },
  productName: {
    fontSize: 22,
    fontWeight: '800',
    color: '#111827',
    lineHeight: 28,
  },
  ratingRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  ratingText: {
    fontSize: 14,
    color: '#6B7280',
  },
  description: {
    fontSize: 14,
    color: '#4B5563',
    lineHeight: 22,
  },
  variantsSection: {
    gap: 8,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#111827',
  },
  variantsScroll: {
    paddingBottom: 4,
    gap: 10,
  },
  variantChip: {
    paddingHorizontal: 14,
    paddingVertical: 10,
    borderRadius: 12,
    backgroundColor: '#FFFFFF',
    borderWidth: 2,
    borderColor: '#E5E7EB',
    minWidth: 90,
    marginRight: 10,
    alignItems: 'center',
  },
  variantChipActive: {
    borderColor: '#1E90FF',
    backgroundColor: '#EBF5FF',
  },
  variantChipOos: {
    opacity: 0.5,
  },
  variantPrice: {
    fontSize: 14,
    fontWeight: '700',
    color: '#111827',
  },
  variantPriceActive: {
    color: '#1E90FF',
  },
  variantAttrs: {
    fontSize: 11,
    color: '#9CA3AF',
    marginTop: 2,
  },
  variantAttrsActive: {
    color: '#1E90FF',
  },
  oosLabel: {
    fontSize: 10,
    color: '#EF4444',
    marginTop: 2,
  },
  addButton: {
    borderRadius: 12,
    marginTop: 4,
  },
  addButtonContent: {
    height: 52,
  },
  addButtonLabel: {
    fontSize: 16,
    fontWeight: '700',
  },
  divider: {
    marginVertical: 8,
    backgroundColor: '#E5E7EB',
  },
  reviewsSection: {
    gap: 10,
  },
  noReviews: {
    fontSize: 14,
    color: '#9CA3AF',
    textAlign: 'center',
    paddingVertical: 16,
  },
  reviewCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 10,
    padding: 12,
    gap: 4,
    borderWidth: 1,
    borderColor: '#F3F4F6',
  },
  reviewHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 4,
  },
  reviewUser: {
    fontSize: 12,
    color: '#9CA3AF',
    fontFamily: 'monospace',
  },
  reviewTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#111827',
  },
  reviewBody: {
    fontSize: 13,
    color: '#4B5563',
    lineHeight: 20,
  },
  snackbar: {
    backgroundColor: '#1F2937',
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
