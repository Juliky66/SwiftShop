import { catalogClient } from './client'
import { Category, Product, Page } from '../types'

export async function getCategories(): Promise<Category[]> {
  const { data } = await catalogClient.get<Category[]>('/api/v1/categories')
  return data
}

export interface ProductsParams {
  limit?: number
  offset?: number
}

export async function getProducts(params: ProductsParams = {}): Promise<Page<Product>> {
  const { data } = await catalogClient.get<Page<Product>>('/api/v1/products', {
    params: { limit: params.limit ?? 20, offset: params.offset ?? 0 },
  })
  return data
}

export async function searchProducts(
  q: string,
  params: ProductsParams = {},
): Promise<Page<Product>> {
  const { data } = await catalogClient.get<Page<Product>>('/api/v1/products/search', {
    params: { q, limit: params.limit ?? 20, offset: params.offset ?? 0 },
  })
  return data
}

export async function getProduct(slug: string): Promise<Product> {
  const { data } = await catalogClient.get<Product>(`/api/v1/products/${slug}`)
  return data
}

export async function getCategoryProducts(
  slug: string,
  params: ProductsParams = {},
): Promise<Page<Product>> {
  const { data } = await catalogClient.get<Page<Product>>(
    `/api/v1/categories/${slug}/products`,
    { params: { limit: params.limit ?? 20, offset: params.offset ?? 0 } },
  )
  return data
}
