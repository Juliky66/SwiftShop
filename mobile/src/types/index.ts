export interface User {
  id: string
  email: string
  full_name: string
  role: string
  phone?: string
}

export interface Category {
  id: string
  name: string
  slug: string
  parent_id: string | null
  children?: Category[]
}

export interface ProductVariant {
  id: string
  sku: string
  price: number
  stock: number
  attributes: Record<string, string>
}

export interface ProductImage {
  id: string
  url: string
  sort_order: number
}

export interface Product {
  id: string
  name: string
  slug: string
  description: string
  category_id: string
  rating: number
  review_count: number
  variants: ProductVariant[]
  images: ProductImage[]
  primary_image_url?: string
}

export interface CartItem {
  id: string
  cart_id: string
  variant_id: string
  quantity: number
  price_snapshot: number
}

export interface Cart {
  id: string
  user_id: string
}

export interface CartResponse {
  cart: Cart
  items: CartItem[]
}

export interface DeliveryAddress {
  full_name: string
  phone: string
  city: string
  street: string
  postal_code: string
}

export interface OrderItem {
  id: string
  order_id: string
  variant_id: string
  product_name: string
  sku: string
  quantity: number
  unit_price: number
  total_price: number
}

export interface Order {
  id: string
  status: string
  total_amount: number
  shipping_address: string
  created_at: string
  items?: OrderItem[]
}

export interface Review {
  id: string
  user_id: string
  product_id: string
  rating: number
  title?: string
  body?: string
  is_approved: boolean
}

export interface Page<T> {
  items: T[]
  total: number
  limit: number
  offset: number
}
