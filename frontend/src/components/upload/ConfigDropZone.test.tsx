import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ConfigDropZone } from './ConfigDropZone'

describe('ConfigDropZone', () => {
  it('shows browse button and supported types', () => {
    render(<ConfigDropZone onFile={() => {}} />)
    expect(screen.getByText('Browse Files')).toBeInTheDocument()
    expect(screen.getByText(/Supported:/)).toBeInTheDocument()
  })

  it('rejects unsupported file types', () => {
    const onFile = vi.fn()
    render(<ConfigDropZone onFile={onFile} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement

    const file = new File(['content'], 'malware.exe', { type: 'application/octet-stream' })
    fireEvent.change(input, { target: { files: [file] } })

    expect(onFile).not.toHaveBeenCalled()
    expect(screen.getByText(/Unsupported file type/)).toBeInTheDocument()
  })

  it('accepts a .conf file', () => {
    const onFile = vi.fn()
    render(<ConfigDropZone onFile={onFile} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement

    const file = new File(['content'], 'zscaler.conf', { type: 'text/plain' })
    fireEvent.change(input, { target: { files: [file] } })

    expect(onFile).toHaveBeenCalledWith(file)
  })
})
