import React, { useState, useCallback } from 'react'
import {
  View,
  FlatList,
  ScrollView,
  StyleSheet,
  RefreshControl,
  TouchableOpacity,
} from 'react-native'
import { Text, Appbar, Chip } from 'react-native-paper'
import { useQuery } from '@tanstack/react-query'
import { SafeAreaView } from 'react-native-safe-area-context'
import { getCategories, getProducts, getCategoryProducts } from '../../src/api/catalog'
import { Category, Product } from '../../src/types'
import ProductCard from '../../src/components/ProductCard'
import LoadingSpinner from '../../src/components/LoadingSpinner'

export default function HomeScreen() {
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null)
  const [refreshing, setRefreshing] = useState(false)

  const categoriesQuery = useQuery({
    queryKey: ['categories'],
    queryFn: getCategories,
  })

  const productsQuery = useQuery({
    queryKey: ['products', selectedCategory?.slug],
    queryFn: () =>
      selectedCategory
        ? getCategoryProducts(selectedCategory.slug, { limit: 40 })
        : getProducts({ limit: 40 }),
  })

  const onRefresh = useCallback(async () => {
    setRefreshing(true)
    await productsQuery.refetch()
    setRefreshing(false)
  }, [productsQuery])

  const products: Product[] = productsQuery.data?.items ?? []

  // Flatten tree to leaf categories (nodes with no children)
  function flattenLeaves(cats: Category[]): Category[] {
    return cats.flatMap((c) =>
      c.children && c.children.length > 0 ? flattenLeaves(c.children) : [c],
    )
  }
  const leafCategories = flattenLeaves(categoriesQuery.data ?? [])

  const renderProduct = ({ item }: { item: Product }) => (
    <View style={styles.cardWrapper}>
      <ProductCard product={item} />
    </View>
  )

  const renderEmpty = () => (
    <View style={styles.emptyState}>
      <Text style={styles.emptyIcon}>🛍️</Text>
      <Text style={styles.emptyText}>No products found</Text>
      <Text style={styles.emptySubtext}>Check back later for new arrivals</Text>
    </View>
  )

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <Appbar.Header style={styles.appbar}>
        <Appbar.Content
          title="SwiftShop"
          titleStyle={styles.appbarTitle}
        />
      </Appbar.Header>

      {/* Categories horizontal scroll */}
      <View style={styles.categoriesContainer}>
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.categoriesScroll}
        >
          <TouchableOpacity
            style={[
              styles.categoryChip,
              selectedCategory === null && styles.categoryChipActive,
            ]}
            onPress={() => setSelectedCategory(null)}
          >
            <Text
              style={[
                styles.categoryChipText,
                selectedCategory === null && styles.categoryChipTextActive,
              ]}
            >
              All
            </Text>
          </TouchableOpacity>
          {leafCategories.map((cat) => (
            <TouchableOpacity
              key={cat.id}
              style={[
                styles.categoryChip,
                selectedCategory?.id === cat.id && styles.categoryChipActive,
              ]}
              onPress={() =>
                setSelectedCategory(selectedCategory?.id === cat.id ? null : cat)
              }
            >
              <Text
                style={[
                  styles.categoryChipText,
                  selectedCategory?.id === cat.id && styles.categoryChipTextActive,
                ]}
              >
                {cat.name}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      </View>

      {/* Products grid */}
      {productsQuery.isLoading ? (
        <LoadingSpinner message="Loading products..." />
      ) : (
        <FlatList
          data={products}
          renderItem={renderProduct}
          keyExtractor={(item) => item.id}
          numColumns={2}
          contentContainerStyle={styles.productList}
          columnWrapperStyle={styles.columnWrapper}
          ListEmptyComponent={renderEmpty}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={onRefresh}
              colors={['#1E90FF']}
              tintColor="#1E90FF"
            />
          }
          showsVerticalScrollIndicator={false}
        />
      )}
    </SafeAreaView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F8FAFF',
  },
  appbar: {
    backgroundColor: '#1E90FF',
    elevation: 0,
  },
  appbarTitle: {
    color: '#FFFFFF',
    fontWeight: '800',
    fontSize: 22,
  },
  categoriesContainer: {
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E5E7EB',
  },
  categoriesScroll: {
    paddingHorizontal: 12,
    paddingVertical: 10,
    gap: 8,
  },
  categoryChip: {
    paddingHorizontal: 14,
    paddingVertical: 7,
    borderRadius: 20,
    backgroundColor: '#F3F4F6',
    borderWidth: 1,
    borderColor: '#E5E7EB',
    marginRight: 8,
  },
  categoryChipActive: {
    backgroundColor: '#1E90FF',
    borderColor: '#1E90FF',
  },
  categoryChipText: {
    fontSize: 13,
    fontWeight: '500',
    color: '#374151',
  },
  categoryChipTextActive: {
    color: '#FFFFFF',
  },
  productList: {
    padding: 12,
    paddingBottom: 24,
  },
  columnWrapper: {
    justifyContent: 'space-between',
  },
  cardWrapper: {
    flex: 0.5,
  },
  emptyState: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 60,
    gap: 8,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 18,
    fontWeight: '600',
    color: '#374151',
  },
  emptySubtext: {
    fontSize: 14,
    color: '#9CA3AF',
  },
})
