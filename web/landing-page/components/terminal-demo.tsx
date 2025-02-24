"use client"
import { Copy } from "lucide-react";
import { Button } from "./ui/button";
import { useCopyToClipboard } from "usehooks-ts";
import { toast } from 'sonner'
import { useAtomValue } from "jotai";
import { selectionAtom } from "@/lib/stores";

export function TerminalDemo() {
  const choice = useAtomValue(selectionAtom)
  const [_, copyToClipboard] = useCopyToClipboard()
  const apiUrl = `gitignore.lol/api/${choice} > .gitignore`
  return (
    <div className="relative w-full rounded-lg overflow-hidden bg-[#1C1C1C] shadow-2xl border border-gray-800">
      <div className="flex justify-between items-center px-4 py-2 bg-[#2D2D2D] border-b border-gray-800">
        <span className="text-gray-300 font-mono text-sm">Example</span>
        <span className="text-gray-500 font-mono text-sm">Bash</span>
      </div>
      <div className="p-4 font-mono text-sm flex flex-row justify-between items-center">
        <div className="flex items-center gap-2 overflow-x-auto whitespace-nowrap">
          <span className="text-pink-400">curl</span>
          <span className="text-gray-300">{apiUrl}</span>
        </div>
        <Button size='icon' variant={'ghost'} className="text-gray-400" onClick={() => {
          copyToClipboard(`curl ${apiUrl}`).then(() => {
            toast.success('Copied to clipboard!')
          }).catch(() => {
            toast.error('Failed to copy to clipboard')
          })
        }}>
          <Copy className="size-4" />
        </Button>
      </div>
    </div>
  )
}

