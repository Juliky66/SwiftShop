import { reviewsClient } from './client'
import { Review, Page } from '../types'

export interface ReviewsParams {
  rating?: number
  limit?: number
  offset?: number
}

export async function getReviews(
  productId: string,
  params: ReviewsParams = {},
): Promise<Page<Review>> {
  const query: Record<string, unknown> = {
    limit: params.limit ?? 20,
    offset: params.offset ?? 0,
  }
  if (params.rating !== undefined) query.rating = params.rating
  const { data } = await reviewsClient.get<Page<Review>>(
    `/api/v1/products/${productId}/reviews`,
    { params: query },
  )
  return data
}

export interface SubmitReviewPayload {
  rating: number
  title?: string
  body?: string
}

export async function submitReview(
  productId: string,
  payload: SubmitReviewPayload,
): Promise<Review> {
  const { data } = await reviewsClient.post<Review>(
    `/api/v1/products/${productId}/reviews`,
    payload,
  )
  return data
}

export interface PaymentResult {
  payment_id: string
  confirm_url: string
  status: string
}

export async function initiatePayment(orderId: string): Promise<PaymentResult> {
  const { data } = await reviewsClient.post<PaymentResult>(`/api/v1/orders/${orderId}/pay`)
  return data
}
