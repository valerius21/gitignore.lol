'use client'
import { MultiSelect } from "@/components/ui/multi-select";
import { selectionAtom } from "@/lib/stores";
import { useQuery } from "@tanstack/react-query";
import { setMaxListeners } from "events";
import { useAtom } from "jotai";
import { useEffect, useState } from "react";

export function SearchBox() {
  const [selection, setSelection] = useAtom(selectionAtom)
  const { data, isLoading } = useQuery<{ templates: string[] }>({
    queryKey: ['ignore-list'],
    queryFn: () => fetch('http://localhost:4444/api/list').then(res => res.json()),
    initialData: ({
      templates: ['python', 'jupyternotebooks']
    })
  })

  return (
    <>
      <MultiSelect
        options={data?.templates.map(s => ({
          label: s,
          value: s
        }))}
        onValueChange={setSelection}
        defaultValue={data?.templates.slice(0, 3).map(s => s)}
        placeholder="Select a template ..."
        variant={'inverted'}
        maxCount={100}
      >
      </MultiSelect>
    </>
  )
}
