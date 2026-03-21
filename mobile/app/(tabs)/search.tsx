import React, { useState, useEffect } from 'react'
import { View, FlatList, StyleSheet } from 'react-native'
import { Text, Searchbar } from 'react-native-paper'
import { useQuery } from '@tanstack/react-query'
import { SafeAreaView } from 'react-native-safe-area-context'
import { searchProducts } from '../../src/api/catalog'
import { Product } from '../../src/types'
import ProductCard from '../../src/components/ProductCard'
import LoadingSpinner from '../../src/components/LoadingSpinner'

export default function SearchScreen() {
  const [inputValue, setInputValue] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')

  // Debounce 500ms
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(inputValue.trim())
    }, 500)
    return () => clearTimeout(timer)
  }, [inputValue])

  const searchQuery = useQuery({
    queryKey: ['search', debouncedQuery],
    queryFn: () => searchProducts(debouncedQuery, { limit: 40 }),
    enabled: debouncedQuery.length > 0,
  })

  const products: Product[] = searchQuery.data?.items ?? []

  const renderProduct = ({ item }: { item: Product }) => (
    <View style={styles.cardWrapper}>
      <ProductCard product={item} />
    </View>
  )

  const renderContent = () => {
    if (!debouncedQuery) {
      return (
        <View style={styles.centerState}>
          <Text style={styles.stateIcon}>🔍</Text>
          <Text style={styles.stateText}>Type to search products...</Text>
        </View>
      )
    }

    if (searchQuery.isLoading) {
      return <LoadingSpinner message="Searching..." />
    }

    if (searchQuery.isError) {
      return (
        <View style={styles.centerState}>
          <Text style={styles.stateIcon}>⚠️</Text>
          <Text style={styles.stateText}>Search unavailable</Text>
          <Text style={styles.stateSubtext}>Please try again later</Text>
        </View>
      )
    }

    if (products.length === 0) {
      return (
        <View style={styles.centerState}>
          <Text style={styles.stateIcon}>😕</Text>
          <Text style={styles.stateText}>No results for "{debouncedQuery}"</Text>
          <Text style={styles.stateSubtext}>Try different keywords</Text>
        </View>
      )
    }

    return (
      <FlatList
        data={products}
        renderItem={renderProduct}
        keyExtractor={(item) => item.id}
        numColumns={2}
        contentContainerStyle={styles.productList}
        columnWrapperStyle={styles.columnWrapper}
        showsVerticalScrollIndicator={false}
      />
    )
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.searchContainer}>
        <Searchbar
          placeholder="Search products..."
          onChangeText={setInputValue}
          value={inputValue}
          style={styles.searchBar}
          inputStyle={styles.searchInput}
          iconColor="#1E90FF"
          placeholderTextColor="#9CA3AF"
          autoCapitalize="none"
          autoCorrect={false}
        />
      </View>
      {renderContent()}
    </SafeAreaView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F8FAFF',
  },
  searchContainer: {
    backgroundColor: '#FFFFFF',
    paddingHorizontal: 12,
    paddingVertical: 10,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E7EB',
  },
  searchBar: {
    backgroundColor: '#F3F4F6',
    borderRadius: 12,
    elevation: 0,
  },
  searchInput: {
    fontSize: 15,
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
  centerState: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    paddingHorizontal: 32,
  },
  stateIcon: {
    fontSize: 48,
    marginBottom: 8,
  },
  stateText: {
    fontSize: 17,
    fontWeight: '600',
    color: '#374151',
    textAlign: 'center',
  },
  stateSubtext: {
    fontSize: 14,
    color: '#9CA3AF',
    textAlign: 'center',
  },
})
