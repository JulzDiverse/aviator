package processor_test

import (
	"errors"
	"os"

	. "github.com/JulzDiverse/aviator/processor"

	fakes "github.com/JulzDiverse/aviator/processor/processorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Processor", func() {
	var aviatorYaml string
	var spruceProcessor *fakes.FakeSpruceProcessor
	var flyExecuter *fakes.FakeFlyExecuter

	BeforeEach(func() {
		spruceProcessor = new(fakes.FakeSpruceProcessor)
		flyExecuter = new(fakes.FakeFlyExecuter)
	})

	Context("New", func() {
		Context("aviator.yml validation", func() {
			var processor *Processor

			Context("spruce section", func() {
				It("is able to read all 'with' related properties", func() {
					aviatorYaml = `spruce:
- base: base.yml
  merge:
  - with:
      files:
      - file1.yml
      - file2.yml
    regexp: ".*.(yml)"
    in_dir: /path/
    skip_non_existing: true
  - with_in: path/to/dir/
    except:
    - file2.yml
  to: result.yml`

					var err error
					processor, err = New([]byte(aviatorYaml), spruceProcessor, flyExecuter)
					Expect(err).ToNot(HaveOccurred())

					Expect(len(processor.Aviator.Spruce[0].Merge[0].With.Files)).To(Equal(2))
					Expect(processor.Aviator.Spruce[0].Merge[1].WithIn).To(Equal("path/to/dir/"))
					Expect(len(processor.Aviator.Spruce[0].Merge[1].Except)).To(Equal(1))
					Expect(processor.Aviator.Spruce[0].Merge[0].Regexp).To(Equal(".*.(yml)"))
					Expect(processor.Aviator.Spruce[0].Merge[0].Skip).To(Equal(true))
					Expect(processor.Aviator.Spruce[0].To).To(Equal("result.yml"))

				})

				It("is able to parse all for_each_in related properties", func() {
					aviatorYaml = `spruce:
- base: result.yml
  merge:
  - with_in: another/path/
  for_each_in: path/to/dir/
  except:
  - file2.yml
  regexp: ".*.(yml)"
  to_dir: some/tmp/dir/`

					var err error
					processor, err = New([]byte(aviatorYaml), spruceProcessor, flyExecuter)
					Expect(err).ToNot(HaveOccurred())

					Expect(processor.Aviator.Spruce[0].ForEachIn).To(Equal("path/to/dir/"))
					Expect(len(processor.Aviator.Spruce[0].Except)).To(Equal(1))
					Expect(processor.Aviator.Spruce[0].ToDir).To(Equal("some/tmp/dir/"))
				})

				It("is able to read all 'cherry_pick' and 'skip_eval' properties", func() {
					aviatorYaml = `spruce:
- base: some/tmp/dir/file1.yml
  cherry_pick:
  - one
  - two
  - three
  merge:
  - with_in: path/
  skip_eval: true
  for_each:
  - foo.yml
  - bar.yml
  to_dir: foo/bar/`

					var err error
					processor, err = New([]byte(aviatorYaml), spruceProcessor, flyExecuter)
					Expect(err).ToNot(HaveOccurred())

					Expect(len(processor.Aviator.Spruce[0].ForEach)).To(Equal(2))
					Expect(len(processor.Aviator.Spruce[0].CherryPicks)).To(Equal(3))
					Expect(processor.Aviator.Spruce[0].SkipEval).To(Equal(true))
				})

				It("is able to read all 'walk_through' related properties", func() {
					aviatorYaml = `spruce:
- base: base.yml
  prune:
  - some
  - properties
  merge:
  - with_in: foo/
  walk_through: foo/bar/
  for_all: some/dir/
  copy_parents: true
  enable_matching: true
  to_dir: final/`

					var err error
					processor, err = New([]byte(aviatorYaml), spruceProcessor, flyExecuter)
					Expect(err).ToNot(HaveOccurred())

					Expect(processor.Aviator.Spruce[0].WalkThrough).To(Equal("foo/bar/"))
					Expect(len(processor.Aviator.Spruce[0].Prune)).To(Equal(2))
					Expect(processor.Aviator.Spruce[0].CopyParents).To(Equal(true))
					Expect(processor.Aviator.Spruce[0].EnableMatching).To(Equal(true))
					Expect(processor.Aviator.Spruce[0].ForAll).To(Equal("some/dir/"))
				})

				It("is able resolve environment variables", func() {
					os.Setenv("ENV_VAR", "envVar")
					os.Setenv("ANOTHER_VAR", "another")
					os.Setenv("RESULT", "result")
					aviatorYaml = `spruce:
- base: $ENV_VAR
  merge:
  - with:
      files:
      - $ANOTHER_VAR
  to: $RESULT`

					var err error
					processor, err = New([]byte(aviatorYaml), spruceProcessor, flyExecuter)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor.Aviator.Spruce[0].Base).To(Equal("envVar"))
					Expect(processor.Aviator.Spruce[0].Merge[0].With.Files[0]).To(Equal("another"))
					Expect(processor.Aviator.Spruce[0].To).To(Equal("result"))
				})

				It("is able to parse '{{}}'", func() {
					os.Setenv("ENV_VAR", "envVar")
					os.Setenv("ANOTHER_VAR", "another")
					os.Setenv("RESULT", "result")
					aviatorYaml = `spruce:
- base: input.yml 
  merge:
  - with:
      files:
      - {{identifier}}
  to: {{result}}`

					var err error
					processor, err = New([]byte(aviatorYaml), spruceProcessor, flyExecuter)
					Expect(err).ToNot(HaveOccurred())

				})
			})

			Context("fly section", func() {
				BeforeEach(func() {
					aviatorYaml = `fly:
  name: pipelineName
  target: targetName
  config: configFile
  expose: true
  vars:
  - credentials.yml`
				})

				It("is able to read all properties from the fly section", func() {
					processor, err := New([]byte(aviatorYaml), spruceProcessor, flyExecuter)
					Expect(err).ToNot(HaveOccurred())

					Expect(processor.Aviator.Fly.Name).To(Equal("pipelineName"))
					Expect(processor.Aviator.Fly.Target).To(Equal("targetName"))
					Expect(processor.Aviator.Fly.Config).To(Equal("configFile"))
					Expect(processor.Aviator.Fly.Expose).To(BeTrue())
					Expect(len(processor.Aviator.Fly.Vars)).To(Equal(1))
				})

				Context("executing fly returns a valid error", func() {
					BeforeEach(func() {
						flyExecuter.ExecuteReturns(errors.New("uups"))
					})
					It("", func() {
						processor, err := New([]byte(aviatorYaml), spruceProcessor, flyExecuter)
						Expect(err).ToNot(HaveOccurred())

						err = processor.ExecuteFly()
						Expect(err).To(MatchError(ContainSubstring("Executing Fly FAILED")))
						Expect(err).To(MatchError(ContainSubstring("uups")))
					})
				})
			})
		})
	})

	Context("spruce section processor", func() {
		BeforeEach(func() {
			aviatorYaml = `spruce:
- base: input.yml 
  merge:
  - with_in: some/dir/
  to: output.yml`
			spruceProcessor.ProcessReturns(nil, errors.New("uups"))
		})

		It("returns a valid error message", func() {
			processor, err := New([]byte(aviatorYaml), spruceProcessor, flyExecuter)
			Expect(err).ToNot(HaveOccurred())
			_, err = processor.ProcessSprucePlan()

			Expect(err).To(MatchError(ContainSubstring("Processing Spruce Plan FAILED")))
			Expect(err).To(MatchError(ContainSubstring("uups")))
		})
	})
})
