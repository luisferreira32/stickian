import React, { useEffect, useRef, useState } from 'react'
import { apiRequest } from '../../shared/auth'

const COLORS = ['royalblue', 'cornflowerblue', 'sandybrown', 'forestgreen', 'dimgray']
const HEX_SIZE = 8

export default function WorldMap() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const mapDataRef = useRef<number[][] | null>(null)
  const transformRef = useRef({ x: 0, y: 0, scale: 1 })
  const isDraggingRef = useRef(false)
  const lastMouseRef = useRef({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)

  useEffect(() => {
    apiRequest('/api/map')
      .then((res) => res.text())
      .then((csv) => {
        const rows = csv.trim().split('\n')
        const data = rows.map((r) => r.split(',').filter(x => x.trim() !== '').map(Number))
        mapDataRef.current = data

        // Initial centering
        if (data.length > 0 && canvasRef.current) {
          const cols = data.length
          const numRows = data[0].length
          const centerQ = cols / 2
          const centerR = numRows / 2
          const targetX = HEX_SIZE * (3 / 2 * centerQ)
          const targetY = HEX_SIZE * Math.sqrt(3) * (centerR + centerQ / 2)
          
          transformRef.current.x = canvasRef.current.width / 2 - targetX
          transformRef.current.y = canvasRef.current.height / 2 - targetY
        }
        
        draw()
      })
      .catch((err) => console.error('Failed to fetch map data', err))
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
    for (let q = 0; q < cols; q++) {
      const rowLen = mapData[q].length
      for (let r = 0; r < rowLen; r++) {
        const val = mapData[q][r]
        if (val === undefined || isNaN(val)) continue

        const x = HEX_SIZE * (3 / 2 * q)
        const y = HEX_SIZE * Math.sqrt(3) * (r + q / 2)

        ctx.fillStyle = COLORS[val] || 'black'
        ctx.beginPath()
        for (let i = 0; i < 6; i++) {
          const angle_deg = 60 * i
          const angle_rad = Math.PI / 180 * angle_deg
          const px = x + HEX_SIZE * 1.05 * Math.cos(angle_rad)
          const py = y + HEX_SIZE * 1.05 * Math.sin(angle_rad)
          if (i === 0) ctx.moveTo(px, py)
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

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!isDraggingRef.current) return
    const dx = e.clientX - lastMouseRef.current.x
    const dy = e.clientY - lastMouseRef.current.y
    transformRef.current.x += dx
    transformRef.current.y += dy
    lastMouseRef.current = { x: e.clientX, y: e.clientY }
    requestAnimationFrame(draw)
  }

  const handleMouseUp = () => {
    isDraggingRef.current = false
    setIsDragging(false)
  }

  return (
    <div style={{ width: '100%', height: 'calc(100vh - 80px)', overflow: 'hidden', padding: 0 }}>
      <canvas
        ref={canvasRef}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        style={{ cursor: isDragging ? 'grabbing' : 'grab', display: 'block', touchAction: 'none' }}
      />
    </div>
  )
}
