import { useEffect, useState } from 'react'
import './City.css'

type CityData = {
    cityName: string
    buildings: Record<string, number>
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
                console.log('City JSON', data)
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

            {Object.entries(city.buildings).map(([name, level]) => (
                <div key={name} className={`building ${name}`}>
                    <div className="building-name">{prettyName(name)}</div>
                    <div className="building-level">(Level {level})</div>
                </div>
            ))}
        </div>
    )
}

export default City