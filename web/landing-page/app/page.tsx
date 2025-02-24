import { SearchBox } from "@/components/search-box"
import { TerminalDemo } from "@/components/terminal-demo"
import { Button } from "@/components/ui/button"
import { Github, FileJson, Search, FileJson2 } from "lucide-react"
import Link from "next/link"

export default function Home() {
  return (
    <div className="min-h-screen bg-black text-white">
      <header className="container mx-auto px-4 py-6 flex justify-between items-center">
        <Link href="/" className="text-xl font-bold font-mono">
          gitignore.lol
        </Link>
        <nav className="flex items-center gap-4">
          <Link href="/documentation" className="text-gray-400 hover:text-white">
            <Button size="sm">
              Documentation
            </Button>
          </Link>
          <Link href="/swagger/index.html" className="text-gray-400 hover:text-white">
            <Button size="sm">
              Swagger / OpenAPI
            </Button>
          </Link>
          <Link
            href="https://github.com/valerius21/gitignore.lol"
            target="_blank"
            rel="noopener noreferrer"
            className="text-gray-400 hover:text-white"
          >
            <Button variant="ghost" size="sm">
              <Github className="h-5 w-5" />
              <span className="sr-only">GitHub</span>
            </Button>
          </Link>
        </nav>
      </header>

      <main className="container mx-auto px-4">
        <div className="max-w-3xl mx-auto space-y-12 py-12">
          <div className="text-center space-y-4">
            <h1 className="text-4xl sm:text-5xl md:text-6xl font-bold tracking-tight bg-gradient-to-r from-purple-400 to-pink-600 text-transparent bg-clip-text p-4 font-mono">
              gitignore.lol
            </h1>
            <p className="text-xl sm:text-2xl text-gray-400">For devs who hate commit noise</p>
            <p className="text-gray-500">No redirects. No ads. Just clean,
              <a href="https://github.com/github/gitignore" target="_blank" className="mx-1 underline underline-offset-2 hover:text-white transition-colors">
                GitHub-powered
              </a>
              templates.</p>
            <div className="pt-4 flex flex-row gap-4 items-center justify-center">
              <Button
                asChild
                size="lg"
                className="bg-gradient-to-r from-purple-600 to-pink-600 hover:from-purple-500 hover:to-pink-500"
              >
                <Link href="https://github.com/github/gitignore" target="_blank">
                  <Search className="mr-2 h-5 w-5" />
                  Browse Templates
                </Link>
              </Button>
              <Button
                size="lg"
                asChild
                className="bg-zinc-800 hover:bg-zinc-800/75"
              >
                <Link href="/swagger/index.html">
                  <FileJson2 className="mr-2 h-5 w-5" />
                  Swagger / OpenAPI
                </Link>
              </Button>
            </div>
          </div>

          <SearchBox />
          <TerminalDemo />

          <div className="grid md:grid-cols-3 gap-6">
            {features.map((feature, index) => (
              <div key={index} className="p-6 rounded-lg border border-gray-800 bg-gray-900/50 space-y-4">
                <div className="text-2xl">{feature.icon}</div>
                <h3 className="text-lg font-semibold">{feature.title}</h3>
                <p className="text-gray-400">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </main>

      <footer className="absolute bottom-0 w-full py-6 text-center text-gray-500 text-sm">
        Built with ♥️ for developers by <a href="https://valerius.me" target="_blank" className="underline">Valerius Mattfeld</a>.
        <a href="https://valerius.me/a6ea07d7-93e9-4377-9630-b3f79779f422" className="ml-2 underline">
          Legal.
        </a>
      </footer>
    </div>
  )
}

const features = [
  {
    icon: "🔗",
    title: "No redirects",
    description: "No weird rebranding and redirects. The base API URL stays the same.",
  },
  {
    icon: "🚀",
    title: "Zero fuss",
    description: "Quick and simple .gitignore generation powered by GitHub's official templates.",
  },
  {
    icon: "💻",
    title: "Web or CLI",
    description: "Generate templates through the web interface or use the REST API - whatever fits your workflow best.",
  },
]


