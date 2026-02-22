import { useState, useEffect, useCallback } from 'react'

interface GameState {
  foo: number
  bar: number
}

const Dummy = () => {
  const [gameState, setGameState] = useState<GameState>({ foo: 0, bar: 0 })
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Fetch current game state from server
  const fetchGameState = useCallback(async () => {
    try {
      const response = await fetch('/api/foobar')
      if (!response.ok) {
        throw new Error(`Failed to fetch game state: ${response.statusText}`)
      }
      const data = await response.json()
      setGameState({ foo: data.foo, bar: data.bar })
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error occurred')
      console.error('Failed to fetch game state:', err)
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Train Foo: costs 1 Bar, will produce 1 Foo after 3 ticks
  const trainFoo = async () => {
    if (gameState.bar < 1) {
      setError('Not enough Bar to train Foo')
      return
    }

    // Optimistic update: subtract 1 Bar immediately
    setGameState((prev) => ({ ...prev, bar: prev.bar - 1 }))

    try {
      const response = await fetch('/api/foo', { method: 'POST' })
      if (!response.ok) {
        throw new Error(`Failed to train Foo: ${response.statusText}`)
      }
      setError(null)
    } catch (err) {
      // Revert optimistic update on error
      setGameState((prev) => ({ ...prev, bar: prev.bar + 1 }))
      setError(err instanceof Error ? err.message : 'Failed to train Foo')
      console.error('Failed to train Foo:', err)
    }
  }

  // Build Bar: costs 1 Bar, produces 2 Bar (net +1 Bar)
  const buildBar = async () => {
    if (gameState.bar < 2) {
      setError('Not enough Bar to build Bar')
      return
    }

    // Optimistic update: net +1 Bar (costs 1, produces 2)
    setGameState((prev) => ({ ...prev, bar: prev.bar + 1 }))

    try {
      const response = await fetch('/api/bar', { method: 'POST' })
      if (!response.ok) {
        throw new Error(`Failed to build Bar: ${response.statusText}`)
      }
      setError(null)
    } catch (err) {
      // Revert optimistic update on error
      setGameState((prev) => ({ ...prev, bar: prev.bar - 1 }))
      setError(err instanceof Error ? err.message : 'Failed to build Bar')
      console.error('Failed to build Bar:', err)
    }
  }

  // Simulate baseline Bar production (1 Bar per second locally for smoother UX)
  useEffect(() => {
    const interval = setInterval(() => {
      setGameState((prev) => ({ ...prev, bar: prev.bar + 1 }))
    }, 1000) // 1 Bar per second

    return () => clearInterval(interval)
  }, [])

  // Fetch actual state from server every 5 seconds to sync
  useEffect(() => {
    // Initial fetch
    fetchGameState()

    // Periodic refetch every 5 seconds
    const interval = setInterval(fetchGameState, 5000)

    return () => clearInterval(interval)
  }, [fetchGameState])

  if (isLoading) {
    return (
      <div className="app">
        <h1>Stickian Game</h1>
        <p>Loading game state...</p>
      </div>
    )
  }

  return (
    <div className="app">
      <h1>Stickian Game</h1>

      {error && (
        <div className="error" style={{ color: 'red', marginBottom: '1rem' }}>
          Error: {error}
        </div>
      )}

      <div className="game-state" style={{ marginBottom: '2rem' }}>
        <div className="resource" style={{ marginBottom: '0.5rem' }}>
          <strong>Foo: {gameState.foo}</strong>
        </div>
        <div className="resource">
          <strong>Bar: {gameState.bar}</strong>
        </div>
      </div>

      <div className="actions">
        <button
          onClick={trainFoo}
          disabled={gameState.bar < 1}
          style={{
            marginRight: '1rem',
            padding: '0.5rem 1rem',
            backgroundColor: gameState.bar >= 1 ? '#4CAF50' : '#ccc',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: gameState.bar >= 1 ? 'pointer' : 'not-allowed',
          }}
        >
          Train Foo (Cost: 1 Bar)
        </button>

        <button
          onClick={buildBar}
          disabled={gameState.bar < 2}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: gameState.bar >= 1 ? '#2196F3' : '#ccc',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: gameState.bar >= 1 ? 'pointer' : 'not-allowed',
          }}
        >
          Build Bar (Cost: 1 Bar, Produces: 2 Bar)
        </button>
      </div>

      <div style={{ marginTop: '2rem', fontSize: '0.9rem', color: '#666' }}>
        <p>• Foo training takes 3 ticks to complete</p>
        <p>• Bar production gives +1 Bar per tick baseline</p>
        <p>• State syncs with server every 5 seconds</p>
      </div>
    </div>
  )
}

export default Dummy
