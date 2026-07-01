'use client'

import { useEffect, useState, type ReactNode } from 'react'

export function MockProvider({ children }: { children: ReactNode }) {
  // If we are not mocking, we are immediately ready.
  // This ensures the server renders the children instantly in production, fixing the SSR bug.
  const [isReady, setIsReady] = useState(
    process.env.NEXT_PUBLIC_API_MODE !== 'mock'
  )

  useEffect(() => {
    async function enableApiMocking() {
      if (process.env.NEXT_PUBLIC_API_MODE === 'mock') {
        const { worker } = await import('@/lib/mocks/browser')
        await worker.start({
          onUnhandledRequest: 'bypass',
        })
        setIsReady(true)
      }
    }

    enableApiMocking()
  }, [])

  if (!isReady) {
    return null
  }

  return <>{children}</>
}
