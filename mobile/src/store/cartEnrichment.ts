import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import AsyncStorage from '@react-native-async-storage/async-storage'

interface CartProductInfo {
  name: string
  slug: string
}

interface CartEnrichmentStore {
  items: Record<string, CartProductInfo>
  setItem: (variantId: string, info: CartProductInfo) => void
}

export const useCartEnrichment = create<CartEnrichmentStore>()(
  persist(
    (set) => ({
      items: {},
      setItem: (variantId, info) =>
        set((state) => ({ items: { ...state.items, [variantId]: info } })),
    }),
    {
      name: 'cart-enrichment',
      storage: createJSONStorage(() => AsyncStorage),
    },
  ),
)
