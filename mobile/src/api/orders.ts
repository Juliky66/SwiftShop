import { ordersClient } from './client'
import { CartResponse, DeliveryAddress, Order, Page } from '../types'

export async function getCart(): Promise<CartResponse> {
  const { data } = await ordersClient.get<CartResponse>('/api/v1/cart')
  return data
}

export async function addToCart(variantId: string, quantity: number): Promise<void> {
  await ordersClient.post('/api/v1/cart/items', {
    variant_id: variantId,
    quantity,
  })
}

export async function updateCartItem(variantId: string, quantity: number): Promise<void> {
  await ordersClient.put(`/api/v1/cart/items/${variantId}`, { quantity })
}

export async function removeCartItem(variantId: string): Promise<void> {
  await ordersClient.delete(`/api/v1/cart/items/${variantId}`)
}

export async function checkout(deliveryAddress: DeliveryAddress): Promise<Order> {
  const { data } = await ordersClient.post<Order>('/api/v1/orders', {
    delivery_address: deliveryAddress,
  })
  return data
}

export interface OrdersParams {
  limit?: number
  offset?: number
}

export async function getOrders(params: OrdersParams = {}): Promise<Page<Order>> {
  const { data } = await ordersClient.get<Page<Order>>('/api/v1/orders', {
    params: { limit: params.limit ?? 20, offset: params.offset ?? 0 },
  })
  return data
}

export async function getOrder(id: string): Promise<Order> {
  const { data } = await ordersClient.get<Order>(`/api/v1/orders/${id}`)
  return data
}

export async function cancelOrder(id: string): Promise<Order> {
  const { data } = await ordersClient.post<Order>(`/api/v1/orders/${id}/cancel`)
  return data
}
