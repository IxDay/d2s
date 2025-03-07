project = "d2s"
project_dir = File.dirname(__FILE__)
build_file = :"#{File.join %w[out server]}"
build_dist = :"#{File.join %w[out dist server]}"
color_file = File.join('internal', 'github', 'colors.go')

docker = ENV["DOCKER"] || "docker"

task default: [:build]

desc "Run editorconfig checks"
task :"test:editorconfig" do
  sh "ec"
end

desc "Run vulnerability checks"
task "test:vulnerability": %i[generate] do
  sh "govulncheck ./..."
end

desc "Run all unit tests"
task test: %i[test:editorconfig test:vulnerability]

directory "out"
directory File.join %w[out dist]

desc "Build the binary of the project"
task build: [:generate, build_file]

desc "Build the binary for distribution of the project"
task :"build:dist" => [build_dist]

file build_file => ["main.go", "out", "_templ.go"] do |t|
  sh "go build -ldflags '-s -w' -o #{t.name} #{t.prerequisites.first}"
end

file build_dist => ["main.go", "_templ.go", color_file] do |t|
  sh "go build -ldflags '-s -w -X main.DefaultConfigPath=/etc/d2s/base.toml'" +
    " -o #{t.name} #{t.prerequisites.first}"
end

desc "Build the docker image"
task :"build:image" do
  sh "#{docker} build -t #{project} ."
end

desc "Generate template files"
task generate: ["_templ.go"]

rule "_templ.go" => ".templ" do |t|
  sh "templ generate -f #{t.prerequisites.first}"
end

desc "Generate github language - color association"
task :"generate:colors" => [color_file]
file color_file do |t|
  puts "code generate #{color_file}"
  https = SimpleHttp.new('https', 'raw.githubusercontent.com')
  resp = https.get('/github-linguist/linguist/refs/heads/main/lib/linguist/languages.yml')
  cb = ->(entry) { '"' + entry.first + '": "' + entry.last['color'] + '"' rescue nil }
  File.open(t.name, File::CREAT | File::TRUNC | File::WRONLY).write(<<~EOF
      package github

      var Colors = map[string]string{
      \t#{YAML.load(resp.body).map(&cb).compact.join(",\n\t") + ","}
      }
    EOF
  )
  puts "gofmt -l -w #{color_file}"
  `gofmt -l -w #{color_file}`
end

desc "Watch source code and rebuild/reload"
task :watch do
  sh "air --build.bin #{build_file} --tmp_dir #{File.dirname(build_file.to_s)}"
end

desc "Serve godoc (localhost:6060)"
task :doc do
  sh "godoc -http=localhost:6060 -play -index -v"
end

desc "Generate table structures in data package"
task :"generate:table" do
  sh "sqddl tables -db './data/d2s.db' -pkg data  -file data/tables.go"
end

desc "Clean up generated files"
task :clean do

  (puts "rm #{color_file}") && File.delete(color_file) if File.exists?(color_file)
  walk("out") do |entry|
    puts "rm #{entry}"
    File.directory?(entry) ? Dir.delete(entry) : File.delete(entry)
  end
  Dir.glob("**/*_templ.go") do |entry|
    puts "rm #{entry}"
    File.delete(entry)
  end
end
