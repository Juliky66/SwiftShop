import React from 'react'
import { View, StyleSheet } from 'react-native'
import { MaterialCommunityIcons } from '@expo/vector-icons'

interface StarRatingProps {
  rating: number
  size?: number
}

export default function StarRating({ rating, size = 16 }: StarRatingProps) {
  const stars = Array.from({ length: 5 }, (_, i) => {
    const starValue = i + 1
    if (rating >= starValue) return 'star'
    if (rating >= starValue - 0.5) return 'star-half-full'
    return 'star-outline'
  }) as Array<'star' | 'star-half-full' | 'star-outline'>

  return (
    <View style={styles.container}>
      {stars.map((iconName, index) => (
        <MaterialCommunityIcons
          key={index}
          name={iconName}
          size={size}
          color="#FFA500"
        />
      ))}
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
  },
})
