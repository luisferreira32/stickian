import { useEffect, useState } from 'react'
import './City.css'

type BuildingQueueItem = {
  building: string
  level: number
  completeTime: string
}

type CityData = {
  id: string
  cityName: string
  buildings: Record<string, number>
  resources: Record<string, number>
  buildingsQueue: BuildingQueueItem[]
}

const prettyName = (name: string) => {
  return name.replace('_', ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  return date.toLocaleTimeString()
}

const City = () => {
  const [city, setCity] = useState<CityData | null>(null)
  const [upgrading, setUpgrading] = useState<string | null>(null)

  const fetchCity = () => {
    fetch('/api/city', { cache: 'no-store' })
      .then((res) => res.json())
      .then((data) => {
        setCity(data)
      })
      .catch(console.error)
  }

  useEffect(() => {
    fetchCity()
    // Refresh city data every 5 seconds to see queue updates
    const interval = setInterval(fetchCity, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleUpgrade = async (building: string, currentLevel: number) => {
    if (!city) return

    setUpgrading(building)

    // Optimistic update: add to queue immediately
    const optimisticQueueItem: BuildingQueueItem = {
      building: building,
      level: currentLevel + 1,
      completeTime: '', // undefined time for optimistic update
    }

    setCity((prevCity) =>
      prevCity
        ? {
            ...prevCity,
            buildingsQueue: prevCity.buildingsQueue
              ? [...prevCity.buildingsQueue, optimisticQueueItem]
              : [optimisticQueueItem],
          }
        : null
    )

    try {
      const response = await fetch('/api/city/building', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          cityID: city.id,
          building: building,
          level: currentLevel + 1,
        }),
      })

      if (response.ok) {
        // Refresh city data after 1 second to get actual queue state
        setTimeout(() => {
          fetchCity()
        }, 1000)
      } else {
        const error = await response.text()
        alert(`Upgrade failed: ${error}`)
        // Remove optimistic update on failure
        fetchCity()
      }
    } catch (error) {
      console.error('Upgrade failed:', error)
      alert('Upgrade failed: Network error')
      // Remove optimistic update on failure
      fetchCity()
    } finally {
      setUpgrading(null)
    }
  }

  if (!city) {
    return <div>Loading City...</div>
  }

  return (
    <div className="city">
      <div className="cityName">{city.cityName}</div>

      <div className="resourceBar">
        {Object.entries(city.resources).map(([name, value]) => (
          <div key={name} className={`resource ${name}`}>
            <span className="resource-name">{prettyName(name)}</span>
            <div className="resource-value">{value}</div>
          </div>
        ))}
      </div>

      {/* Always show building queue with empty slots */}
      <div className="buildingsQueue">
        <h3>Building Queue</h3>
        {Array.from(
          { length: Math.max(2, city.buildingsQueue?.length) },
          (_, index) => {
            const item = city.buildingsQueue[index]
            if (item) {
              return (
                <div key={index} className="queueItem">
                  <span className="queueBuilding">
                    {prettyName(item.building)} → Level {item.level}
                  </span>
                  <span className="queueTime">
                    {item.completeTime
                      ? `Complete: ${formatTime(item.completeTime)}`
                      : 'Processing...'}
                  </span>
                </div>
              )
            } else {
              return (
                <div key={index} className="queueItem emptySlot">
                  <span className="queueBuilding">Empty slot</span>
                  <span className="queueTime">-</span>
                </div>
              )
            }
          }
        )}
      </div>

      <div className="buildings">
        <h3>Buildings</h3>
        {Object.entries(city.buildings).map(([name, level]) => (
          <div key={name} className={`building ${name}`}>
            <div className="buildingInfo">
              <div className="building-name">{prettyName(name)}</div>
              <div className="building-level">Level {level}</div>
            </div>
            <button
              className="upgradeButton"
              onClick={() => handleUpgrade(name, level)}
              disabled={upgrading === name}
            >
              {upgrading === name ? 'Upgrading...' : `Upgrade to ${level + 1}`}
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}

export default City
