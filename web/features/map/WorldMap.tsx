import React, { useEffect, useRef, useState, ChangeEvent } from 'react'
import { apiRequest } from '../../shared/auth'
import './WorldMap.css'

const BIOME_COLORS: Record<number, string> = {
  0: 'royalblue',      // ocean
  1: 'cornflowerblue', // sea
  2: 'sandybrown',     // beach
  3: 'forestgreen',    // plains
  4: 'dimgray'         // mountain
}
const HEX_SIZE = 8

interface Coords {
  minQ: number
  maxQ: number
  minR: number
  maxR: number
}

export default function WorldMap() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const mapDataRef = useRef<number[][] | null>(null)
  const transformRef = useRef({ x: 0, y: 0, scale: 1 })
  const isDraggingRef = useRef(false)
  const lastMouseRef = useRef({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)

  const [coords, setCoords] = useState<Coords>({
    minQ: 0,
    maxQ: 50,
    minR: 0,
    maxR: 50
  })
  const [hoveredTile, setHoveredTile] = useState<{ q: number, r: number } | null>(null)

  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (isDraggingRef.current) {
      const dx = e.clientX - lastMouseRef.current.x
      const dy = e.clientY - lastMouseRef.current.y
      transformRef.current.x += dx
      transformRef.current.y += dy
      lastMouseRef.current = { x: e.clientX, y: e.clientY }
      requestAnimationFrame(draw)
      return
    }

    const canvas = canvasRef.current
    if (!canvas) return
    const rect = canvas.getBoundingClientRect()
    const screenX = e.clientX - rect.left
    const screenY = e.clientY - rect.top

    // Screen to World
    const worldX = (screenX - transformRef.current.x) / transformRef.current.scale
    const worldY = (screenY - transformRef.current.y) / transformRef.current.scale

    // Inverse transform for x = (q+r)*1.5*S, y = (q-r)*sqrt(3)/2*S
    const xn = worldX / (1.5 * HEX_SIZE)
    const yn = worldY / (HEX_SIZE * Math.sqrt(3) / 2)

    const qf = (xn + yn) / 2
    const rf = (xn - yn) / 2

    // Hex rounding
    let qi = Math.round(qf)
    let ri = Math.round(rf)
    let si = Math.round(-qf - rf)
    const dq = Math.abs(qi - qf)
    const dr = Math.abs(ri - rf)
    const ds = Math.abs(si - (-qf - rf))

    if (dq > dr && dq > ds) qi = -ri - si
    else if (dr > ds) ri = -qi - si

    if (mapDataRef.current) {
      const i = qi - coords.minQ
      const j = ri - coords.minR
      if (mapDataRef.current[i] && mapDataRef.current[i][j] !== undefined) {
        setHoveredTile({ q: qi, r: ri })
      } else {
        setHoveredTile(null)
      }
    }
  }

  const fetchMap = () => {
    apiRequest('/api/map', {
      method: 'POST',
      body: JSON.stringify({
        MinQ: coords.minQ,
        MaxQ: coords.maxQ,
        MinR: coords.minR,
        MaxR: coords.maxR
      })
    })
      .then((res) => res.json())
      .then((resJson) => {
        const data = resJson.biome
        mapDataRef.current = data

        // Initial centering
        if (data.length > 0 && canvasRef.current) {
          const centerQ = (coords.minQ + coords.maxQ) / 2
          const centerR = (coords.minR + coords.maxR) / 2

          const targetX = (centerQ + centerR) * HEX_SIZE * 1.5
          const targetY = (centerQ - centerR) * HEX_SIZE * Math.sqrt(3) / 2

          transformRef.current.x = canvasRef.current.width / 2 - targetX * transformRef.current.scale
          transformRef.current.y = canvasRef.current.height / 2 - targetY * transformRef.current.scale
        }

        draw()
      })
      .catch((err) => console.error('Failed to fetch map data', err))
  }

  useEffect(() => {
    fetchMap()
  }, [])

  const draw = () => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.fillStyle = '#0d1b2a'
    ctx.fillRect(0, 0, canvas.width, canvas.height)

    const mapData = mapDataRef.current
    if (!mapData) {
      ctx.fillStyle = 'white'
      ctx.font = '20px Arial'
      ctx.fillText('Loading map...', canvas.width / 2 - 60, canvas.height / 2)
      return
    }

    ctx.save()
    ctx.translate(transformRef.current.x, transformRef.current.y)
    ctx.scale(transformRef.current.scale, transformRef.current.scale)

    const cols = mapData.length
    for (let i = 0; i < cols; i++) {
      const q = i + coords.minQ
      const rowLen = mapData[i].length
      for (let j = 0; j < rowLen; j++) {
        const r = j + coords.minR
        const type = mapData[i][j]
        if (type === undefined || isNaN(type)) continue

        ctx.fillStyle = BIOME_COLORS[type]

        const x = (q + r) * HEX_SIZE * 1.5
        const y = (q - r) * HEX_SIZE * Math.sqrt(3) / 2

        ctx.beginPath()
        for (let k = 0; k < 6; k++) {
          const angle_deg = 60 * k
          const angle_rad = Math.PI / 180 * angle_deg
          const px = x + HEX_SIZE * 1.05 * Math.cos(angle_rad)
          const py = y + HEX_SIZE * 1.05 * Math.sin(angle_rad)
          if (k === 0) ctx.moveTo(px, py)
          else ctx.lineTo(px, py)
        }
        ctx.closePath()
        ctx.fill()
      }
    }
    ctx.restore()
  }

  // Handle Resize
  useEffect(() => {
    const resizeObj = new ResizeObserver(() => {
      if (!canvasRef.current) return
      const parent = canvasRef.current.parentElement
      if (parent) {
        canvasRef.current.width = parent.clientWidth
        canvasRef.current.height = parent.clientHeight
        draw()
      }
    })
    if (canvasRef.current?.parentElement) {
      resizeObj.observe(canvasRef.current.parentElement)
    }
    return () => resizeObj.disconnect()
  }, [])

  // Passive wheel handler to prevent default scrolling
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const handleWheel = (e: WheelEvent) => {
      e.preventDefault()
      const zoomSensitivity = 0.001
      const delta = -e.deltaY * zoomSensitivity
      const newScale = Math.min(Math.max(0.1, transformRef.current.scale * Math.exp(delta)), 10)

      const rect = canvas.getBoundingClientRect()
      const mouseX = e.clientX - rect.left
      const mouseY = e.clientY - rect.top

      transformRef.current.x = mouseX - (mouseX - transformRef.current.x) * (newScale / transformRef.current.scale)
      transformRef.current.y = mouseY - (mouseY - transformRef.current.y) * (newScale / transformRef.current.scale)
      transformRef.current.scale = newScale

      requestAnimationFrame(draw)
    }

    canvas.addEventListener('wheel', handleWheel, { passive: false })
    return () => {
      canvas.removeEventListener('wheel', handleWheel)
    }
  }, [])

  const handleMouseDown = (e: React.MouseEvent) => {
    isDraggingRef.current = true
    setIsDragging(true)
    lastMouseRef.current = { x: e.clientX, y: e.clientY }
  }


  const handleMouseUp = () => {
    isDraggingRef.current = false
    setIsDragging(false)
  }

  return (
    <div className="world-map">
      <div className="world-map-controls">
        <div className="world-map-controls-grid">
          <label>Min Q: <input type="number" value={coords.minQ} onChange={(e: ChangeEvent<HTMLInputElement>) => setCoords((prev: Coords) => ({ ...prev, minQ: parseInt(e.target.value) }))} className="world-map-controls-input" /></label>
          <label>Max Q: <input type="number" value={coords.maxQ} onChange={(e: ChangeEvent<HTMLInputElement>) => setCoords((prev: Coords) => ({ ...prev, maxQ: parseInt(e.target.value) }))} className="world-map-controls-input" /></label>
          <label>Min R: <input type="number" value={coords.minR} onChange={(e: ChangeEvent<HTMLInputElement>) => setCoords((prev: Coords) => ({ ...prev, minR: parseInt(e.target.value) }))} className="world-map-controls-input" /></label>
          <label>Max R: <input type="number" value={coords.maxR} onChange={(e: ChangeEvent<HTMLInputElement>) => setCoords((prev: Coords) => ({ ...prev, maxR: parseInt(e.target.value) }))} className="world-map-controls-input" /></label>
        </div>
        <button onClick={fetchMap} className="world-map-controls-button">Fetch Map</button>
      </div>
      {hoveredTile && (
        <div className="world-map-hovered-tile">
          Q: {hoveredTile.q}, R: {hoveredTile.r}
        </div>
      )}
      <canvas
        ref={canvasRef}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        className={isDragging ? 'world-map-canvas-grabbing' : 'world-map-canvas'}
      />
    </div>
  )
}
