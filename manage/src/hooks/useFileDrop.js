import { useState, useEffect, useRef } from 'react'

//OS のファイルドラッグか（dnd-kit による一覧の並べ替えのドラッグとは区別する）
const isFileDrag = (e) => Array.from(e.dataTransfer?.types || []).includes('Files')

//ドロップされたファイルを取り出す。フォルダは除く
const droppedFiles = (dataTransfer) => {
  const items = Array.from(dataTransfer.items || [])
  if (items.length === 0) return Array.from(dataTransfer.files || [])
  return items
    .filter(item => item.kind === 'file' && !item.webkitGetAsEntry?.()?.isDirectory)
    .map(item => item.getAsFile())
    .filter(Boolean)
}

//ファイルのドロップを受け付ける要素に渡すハンドラと、ドラッグ中かどうかを返す。
//enabled が false のときは受け付けない。
//この画面にいるあいだは、枠の外に落としてもブラウザがファイルを開いて画面から離れないようにする。
export function useFileDrop(onFiles, enabled = true) {
  const [dragging, setDragging] = useState(false)
  const depth = useRef(0)

  useEffect(() => {
    const prevent = (e) => { if (isFileDrag(e)) e.preventDefault() }
    window.addEventListener('dragover', prevent)
    window.addEventListener('drop', prevent)
    return () => {
      window.removeEventListener('dragover', prevent)
      window.removeEventListener('drop', prevent)
    }
  }, [])

  useEffect(() => {
    if (!enabled) {
      depth.current = 0
      setDragging(false)
    }
  }, [enabled])

  const accepts = (e) => enabled && isFileDrag(e)

  const handlers = {
    onDragEnter: (e) => {
      if (!accepts(e)) return
      e.preventDefault()
      depth.current += 1
      setDragging(true)
    },
    onDragOver: (e) => {
      if (!accepts(e)) return
      e.preventDefault()
      e.dataTransfer.dropEffect = 'copy'
    },
    onDragLeave: (e) => {
      if (!accepts(e)) return
      depth.current = Math.max(0, depth.current - 1)
      if (depth.current === 0) setDragging(false)
    },
    onDrop: (e) => {
      if (!accepts(e)) return
      e.preventDefault()
      depth.current = 0
      setDragging(false)
      onFiles(droppedFiles(e.dataTransfer))
    },
  }

  return { dragging, handlers }
}
