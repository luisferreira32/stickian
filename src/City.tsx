import { useEffect, useState } from 'react'
import './City.css'

type CityData = {
    cityName: string
    buildings: Record<string, number>
    resources: Record<string, number>
}

function prettyName(name: string) {
  return name
    .replace('_', ' ')
    .replace(/\b\w/g, c => c.toUpperCase())
}

function City() {

    const [city, setCity] = useState<CityData | null>(null)

    useEffect(() => {
        fetch('/api/city', {cache: 'no-store'})
            .then(res => res.json())
            .then(data => {
                setCity(data)
            })
            .catch(console.error)
    }, [])

    if (!city){
        return <div>
            Loading City...
        </div>
    }

    return (
        <div className="city">
            <div className="cityName">
                {city.cityName}
            </div>
            <div className = "resourceBar">
                {Object.entries(city.resources).map(([name, value]) => (
                    <div key={name} className={`resource ${name}`}>
                        <span className="resource-name">{prettyName(name)}</span>
                        <div className="resource-value">{value}</div>
                    </div>
                ))}
            </div>

            <div>
                {Object.entries(city.buildings).map(([name, level]) => (
                    <div key={name} className={`building ${name}`}>
                        <div className="building-name">{prettyName(name)}</div>
                        <div className="building-level">(Level {level})</div>
                    </div>
                ))}
            </div>
        </div>
    )
}

export default City