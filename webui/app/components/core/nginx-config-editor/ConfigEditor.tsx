import { useRef, useEffect, useState } from 'react'
import Editor from '@monaco-editor/react'
import type { editor } from 'monaco-editor'
import { Card, CardContent, CardHeader, CardTitle } from '~/components/ui/card'
import { Button } from '~/components/ui/button'
import { Badge } from '~/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '~/components/ui/tabs'
import { ScrollArea } from '~/components/ui/scroll-area'
import { CheckCircle, XCircle, Code, Download } from 'lucide-react'

interface ValidationResult {
  is_valid: boolean
  errors: string[]
  output: string
}

interface ConfigEditorProps {
  value: string
  onChange: (value: string) => void
  validation?: ValidationResult | null
  onValidate?: (content: string) => void
  readOnly?: boolean
  height?: string
  showPreview?: boolean
  enableAutocompletion?: boolean
  theme?: 'light' | 'dark'
}

export default function ConfigEditor({
  value,
  onChange,
  validation,
  onValidate,
  readOnly = false,
  height = '400px',
  showPreview = true,
  enableAutocompletion = true,
  theme = 'light'
}: ConfigEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | undefined>(undefined)
  const [isEditorReady, setIsEditorReady] = useState(false)

  const handleEditorDidMount = (editor: editor.IStandaloneCodeEditor, monaco: typeof import('monaco-editor')) => {
    editorRef.current = editor
    setIsEditorReady(true)
  }

  const handleEditorChange = (value: string | undefined) => {
    const newValue = value || ''
    onChange(newValue)

    // Trigger validation after a delay
    if (onValidate && newValue.trim()) {
      setTimeout(() => onValidate(newValue), 1000)
    }
  }

  const downloadConfig = () => {
    const blob = new Blob([value], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'nginx.conf'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <Code className="h-4 w-4" />
          <span className="font-medium">Configuration Editor</span>
          {validation && (
            <Badge variant={validation.is_valid ? "default" : "destructive"} className="ml-2">
              {validation.is_valid ? "Valid" : "Invalid"}
            </Badge>
          )}
        </div>

        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={downloadConfig}
            disabled={!value.trim()}
          >
            <Download className="h-4 w-4 mr-2" />
            Download
          </Button>
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          <Tabs defaultValue="editor" className="w-full">
            <div className="border-b px-4 py-2">
              <TabsList className="h-9">
                <TabsTrigger value="editor" className="h-7">Editor</TabsTrigger>
                {showPreview && (
                  <TabsTrigger value="preview" className="h-7">Preview</TabsTrigger>
                )}
              </TabsList>
            </div>

            <TabsContent value="editor" className="m-0">
              <div style={{ height }}>
                <Editor
                  language="nginx"
                  theme={theme === 'dark' ? 'vs-dark' : 'vs'}
                  value={value}
                  onChange={handleEditorChange}
                  onMount={handleEditorDidMount}
                  options={{
                    readOnly,
                    automaticLayout: true,
                    scrollBeyondLastLine: false,
                    minimap: { enabled: false },
                    lineNumbers: 'on',
                    glyphMargin: true,
                    folding: true,
                    renderLineHighlight: 'line'
                  }}
                />
              </div>
            </TabsContent>

            {showPreview && (
              <TabsContent value="preview" className="m-0">
                <div className="p-4" style={{ height }}>
                  <ScrollArea className="h-full w-full">
                    <pre className="text-sm text-gray-700 whitespace-pre-wrap">
                      {value || 'No configuration content to preview'}
                    </pre>
                  </ScrollArea>
                </div>
              </TabsContent>
            )}
          </Tabs>
        </CardContent>
      </Card>

      {validation && (
        <Card className="mt-4">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Validation Results</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center space-x-2">
                {validation.is_valid ? (
                  <>
                    <CheckCircle className="h-4 w-4 text-green-500" />
                    <Badge variant="default" className="bg-green-100 text-green-800">
                      Valid Configuration
                    </Badge>
                  </>
                ) : (
                  <>
                    <XCircle className="h-4 w-4 text-red-500" />
                    <Badge variant="destructive">
                      Invalid Configuration
                    </Badge>
                  </>
                )}
              </div>

              {validation.errors.length > 0 && (
                <div className="mt-2">
                  <h4 className="text-sm font-medium text-red-700 mb-2">Errors:</h4>
                  <ScrollArea className="h-32 w-full rounded border p-2">
                    {validation.errors.map((error, index) => (
                      <div key={index} className="text-sm text-red-600 mb-1">
                        {error}
                      </div>
                    ))}
                  </ScrollArea>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
