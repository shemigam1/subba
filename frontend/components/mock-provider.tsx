'use client'

import { useEffect, useState, type ReactNode } from 'react'

export function MockProvider({ children }: { children: ReactNode }) {
  const isMock = process.env.NEXT_PUBLIC_API_MODE === 'mock'
  // If we are not mocking, we are immediately ready.
  const [isReady, setIsReady] = useState(!isMock)

  useEffect(() => {
    async function enableApiMocking() {
      if (!isMock) return

      try {
        const { worker } = await import('@/lib/mocks/browser')
        await worker.start({
          onUnhandledRequest: 'bypass',
        })
        console.log("MSW Worker Started Successfully");
      } catch (e) {
        console.error("MSW Worker failed to start:", e);
      } finally {
        setIsReady(true)
      }
    }

    enableApiMocking()
  }, [isMock])

  if (!isReady) {
    return null
  }

  return <>{children}</>
}
