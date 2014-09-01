// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/juju/cmd"
	gitjujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "launchpad.net/gocheck"

	"github.com/juju/juju/testing"
)

type PluginSuite struct {
	testing.FakeJujuHomeSuite
	oldPath string
}

var _ = gc.Suite(&PluginSuite{})

func (suite *PluginSuite) SetUpTest(c *gc.C) {
	suite.FakeJujuHomeSuite.SetUpTest(c)
	suite.oldPath = os.Getenv("PATH")
	if runtime.GOOS != "windows" {
		os.Setenv("PATH", "/bin:"+gitjujutesting.HomePath())
	} else {
		os.Setenv("PATH", gitjujutesting.HomePath())
	}
}

func (suite *PluginSuite) TearDownTest(c *gc.C) {
	os.Setenv("PATH", suite.oldPath)
	suite.FakeJujuHomeSuite.TearDownTest(c)
}

func (*PluginSuite) TestFindPlugins(c *gc.C) {
	plugins := findPlugins()
	c.Assert(plugins, gc.DeepEquals, []string{})
}

func list_dir(dir string) {
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		fmt.Println(f.Name())
	}
}

func (suite *PluginSuite) TestFindPluginsOrder(c *gc.C) {
	var compare []string
	if runtime.GOOS != "windows" {
		suite.makePlugin("foo", 0744)
		suite.makePlugin("bar", 0654)
		suite.makePlugin("baz", 0645)
		compare = append(compare, "juju-bar", "juju-baz", "juju-foo")
	} else {
		suite.makePluginWindows("foo")
		suite.makePluginWindows("bar")
		suite.makePluginWindows("baz")
		compare = append(compare, "juju-bar.bat", "juju-baz.bat", "juju-foo.bat")
	}
	plugins := findPlugins()
	c.Assert(plugins, gc.DeepEquals, compare)

}

func (suite *PluginSuite) TestFindPluginsIgnoreNotExec(c *gc.C) {
	if runtime.GOOS == "windows" {
		c.Skip("Different permissions on Windows!")
	}
	suite.makePlugin("foo", 0644)
	suite.makePlugin("bar", 0666)
	plugins := findPlugins()
	c.Assert(plugins, gc.DeepEquals, []string{})
}

func (suite *PluginSuite) TestRunPluginExising(c *gc.C) {
	if runtime.GOOS == "windows" {
		suite.makePluginWindows("foo")
	} else {
		suite.makePlugin("foo", 0755)

	}
	ctx := testing.Context(c)
	err := RunPlugin(ctx, "foo.bat", []string{"some params"})
	c.Assert(err, gc.IsNil)
	c.Assert(testing.Stdout(ctx), gc.Equals, "foo some params\n")
	c.Assert(testing.Stderr(ctx), gc.Equals, "")
}

func (suite *PluginSuite) TestRunPluginWithFailing(c *gc.C) {
	failMessage := "failing\n"
	if runtime.GOOS != "windows" {
		suite.makeFailingPlugin("foo", 2)
	} else {
		suite.makeFailingPluginWindows("foo", 2)
		failMessage = strings.Replace(failMessage, "\n", "\r\n", -1)
	}
	ctx := testing.Context(c)
	err := RunPlugin(ctx, "foo", []string{"some params"})
	c.Assert(err, gc.ErrorMatches, "subprocess encountered error code 2")
	c.Assert(err, jc.Satisfies, cmd.IsRcPassthroughError)
	c.Assert(testing.Stdout(ctx), gc.Equals, failMessage)
	c.Assert(testing.Stderr(ctx), gc.Equals, "")
}

func (suite *PluginSuite) TestGatherDescriptionsInParallel(c *gc.C) {
	if runtime.GOOS == "windows" {
		c.Skip("Skip on Windows!")
	}
	// Make plugins that will deadlock if we don't start them in parallel.
	// Each plugin depends on another one being started before they will
	// complete. They make a full loop, so no sequential ordering will ever
	// succeed.
	suite.makeFullPlugin(PluginParams{Name: "foo", Creates: "foo", DependsOn: "bar"})
	suite.makeFullPlugin(PluginParams{Name: "bar", Creates: "bar", DependsOn: "baz"})
	suite.makeFullPlugin(PluginParams{Name: "baz", Creates: "baz", DependsOn: "error"})
	suite.makeFullPlugin(PluginParams{Name: "error", ExitStatus: 1, Creates: "error", DependsOn: "foo"})
	// If the code was wrong, GetPluginDescriptions would deadlock,
	// so timeout after a short while
	resultChan := make(chan []PluginDescription)
	go func() {
		resultChan <- GetPluginDescriptions()
	}()
	// 10 seconds is arbitrary but should always be generously long. Test
	// actually only takes about 15ms in practice. But 10s allows for system hiccups, etc.
	waitTime := 10 * time.Second
	var results []PluginDescription
	select {
	case results = <-resultChan:
		break
	case <-time.After(waitTime):
		c.Fatalf("took longer than %fs to complete.", waitTime.Seconds())
	}

	c.Assert(results, gc.HasLen, 4)
	c.Assert(results[0].name, gc.Equals, "bar")
	c.Assert(results[0].description, gc.Equals, "bar description")
	c.Assert(results[1].name, gc.Equals, "baz")
	c.Assert(results[1].description, gc.Equals, "baz description")
	c.Assert(results[2].name, gc.Equals, "error")
	c.Assert(results[2].description, gc.Equals, "error occurred running 'juju-error --description'")
	c.Assert(results[3].name, gc.Equals, "foo")
	c.Assert(results[3].description, gc.Equals, "foo description")
}

func (suite *PluginSuite) TestHelpPluginsWithNoPlugins(c *gc.C) {
	output := badrun(c, 0, "help", "plugins")
	c.Assert(output, jc.HasPrefix, PluginTopicText)
	c.Assert(output, jc.HasSuffix, "\n\nNo plugins found.\n")
}

func (suite *PluginSuite) TestHelpPluginsWithPlugins(c *gc.C) {
	suite.makeFullPlugin(PluginParams{Name: "foo"})
	suite.makeFullPlugin(PluginParams{Name: "bar"})
	output := badrun(c, 0, "help", "plugins")
	c.Assert(output, jc.HasPrefix, PluginTopicText)
	expectedPlugins := `

bar  bar description
foo  foo description
`

	if runtime.GOOS == "windows" {
		expectedPlugins = `

bar.bat  bar description
foo.bat  foo description
`
		output = strings.Replace(output, "\r\n", "\n", -1)
		output = strings.Replace(output, "\"", "", -1)
	}
	c.Assert(output, jc.HasSuffix, expectedPlugins)
}

func (suite *PluginSuite) TestHelpPluginName(c *gc.C) {
	suite.makeFullPlugin(PluginParams{Name: "foo"})
	output := badrun(c, 0, "help", "foo")
	expectedHelp := `foo longer help

something useful
`
	if runtime.GOOS == "windows" {
		expectedHelp = strings.Replace(expectedHelp, "\n", "\r\n", -1)
		output = strings.Replace(output, "\"", "", -1)
	}
	c.Assert(output, gc.Matches, expectedHelp)
}

func (suite *PluginSuite) TestHelpPluginNameNotAPlugin(c *gc.C) {
	output := badrun(c, 0, "help", "foo")
	expectedHelp := "ERROR unknown command or topic for foo\n"
	c.Assert(output, gc.Matches, expectedHelp)
}

func (suite *PluginSuite) TestHelpAsArg(c *gc.C) {
	suite.makeFullPlugin(PluginParams{Name: "foo"})
	output := badrun(c, 0, "foo", "--help")
	expectedHelp := `foo longer help

something useful
`
	if runtime.GOOS == "windows" {
		expectedHelp = strings.Replace(expectedHelp, "\n", "\r\n", -1)
		output = strings.Replace(output, "\"", "", -1)
	}
	c.Assert(output, gc.Matches, expectedHelp)
}

func (suite *PluginSuite) TestDebugAsArg(c *gc.C) {
	suite.makeFullPlugin(PluginParams{Name: "foo"})
	output := badrun(c, 0, "foo", "--debug")
	expectedDebug := "some debug\n"
	if runtime.GOOS == "windows" {
		expectedDebug = strings.Replace(expectedDebug, "\n", "\r\n", -1)
		output = strings.Replace(output, "\"", "", -1)
	}
	c.Assert(output, gc.Matches, expectedDebug)
}

func (suite *PluginSuite) TestJujuEnvVars(c *gc.C) {
	suite.makeFullPlugin(PluginParams{Name: "foo"})
	output := badrun(c, 0, "foo", "-e", "myenv", "-p", "pluginarg")
	expectedDebug := "foo -e myenv -p pluginarg\n.*env is:  myenv\n.*home is: .*\\.juju\n"
	if runtime.GOOS == "windows" {
		expectedDebug = strings.Replace(expectedDebug, "\n", "\r\n", -1)
		output = strings.Replace(output, "\"", "", -1)
	}
	c.Assert(output, gc.Matches, expectedDebug)
}

func (suite *PluginSuite) makePlugin(name string, perm os.FileMode) {
	content := fmt.Sprintf(`#!/bin/bash \necho %s $1 $2 $3 $4`, name)
	filename := gitjujutesting.HomePath(JujuPluginPrefix + name)
	ioutil.WriteFile(filename, []byte(content), perm)
}

func (suite *PluginSuite) makePluginWindows(name string) {
	content := fmt.Sprintf("@echo off \necho %s %*", name)
	filename := gitjujutesting.HomePath(JujuPluginPrefix + name)
	ioutil.WriteFile(filename+".bat", []byte(content), 0755)
}

func (suite *PluginSuite) makeFailingPlugin(name string, exitStatus int) {
	content := fmt.Sprintf("#!/bin/bash --norc\necho failing\nexit %d", exitStatus)
	filename := gitjujutesting.HomePath(JujuPluginPrefix + name)
	ioutil.WriteFile(filename, []byte(content), 0755)
}

func (suite *PluginSuite) makeFailingPluginWindows(name string, exitStatus int) {
	content := fmt.Sprintf("@echo off & echo failing& exit %d", exitStatus)
	filename := gitjujutesting.HomePath(JujuPluginPrefix + name)
	ioutil.WriteFile(filename+".bat", []byte(content), 0755)
}

type PluginParams struct {
	Name       string
	ExitStatus int
	Creates    string
	DependsOn  string
}

const pluginTemplateBash = `#!/bin/bash --norc

if [ "$1" = "--description" ]; then
  if [ -n "{{.Creates}}" ]; then
    touch "{{.Creates}}"
  fi
  if [ -n "{{.DependsOn}}" ]; then
    # Sleep 10ms while waiting to allow other stuff to do work
    while [ ! -e "{{.DependsOn}}" ]; do sleep 0.010; done
  fi
  echo "{{.Name}} description"
  exit {{.ExitStatus}}
fi

if [ "$1" = "--help" ]; then
  echo "{{.Name}} longer help"
  echo ""
  echo "something useful"
  exit {{.ExitStatus}}
fi

if [ "$1" = "--debug" ]; then
  echo "some debug"
  exit {{.ExitStatus}}
fi

echo {{.Name}} $*
echo "env is: " $JUJU_ENV
echo "home is: " $JUJU_HOME
exit {{.ExitStatus}}
`

const pluginTemplateBatch = `@echo off
if [%1] equ [--description] (
  if [{{.Creates}}] neq [] (
  	if not exist "{{.Creates}}" (
    	fsutil file createnew "{{.Creates}}" 0
    	) else (
    		copy /b "{{.Creates}}" +,,
    	)
 	)
  if [{{.DependsOn}}] neq [] (
    # Sleep 1sec while waiting to allow other stuff to do work
    :loop 
    if not exist "{{.DependsOn}}" (
    	 timeout 1
    	 goto loop)
 	)
  echo "{{.Name}} description"
  exit {{.ExitStatus}}
)
if [%1] equ [--help] (
  echo "{{.Name}} longer help"
  echo ""
  echo "something useful"
  exit {{.ExitStatus}}
)
if [%1] equ [--debug] (
  echo "some debug"
  exit {{.ExitStatus}}
)
echo {{.Name}} %*
echo "env is:" %JUJU_ENV%
echo "home is:" %JUJU_HOME%
exit {{.ExitStatus}}
`

func (suite *PluginSuite) makeFullPlugin(params PluginParams) {
	var t *template.Template
	content := &bytes.Buffer{}
	filename := gitjujutesting.HomePath("juju-" + params.Name)

	// Create a new template and parse the plugin into it.
	if runtime.GOOS != "windows" {
		t = template.Must(template.New("plugin").Parse(pluginTemplateBash))
	} else {
		t = template.Must(template.New("plugin").Parse(pluginTemplateBatch))
		filename = filename + ".bat"
	}

	// Create the files in the temp dirs, so we don't pollute the working space
	if params.Creates != "" {
		params.Creates = gitjujutesting.HomePath(params.Creates)
	}
	if params.DependsOn != "" {
		params.DependsOn = gitjujutesting.HomePath(params.DependsOn)
	}
	t.Execute(content, params)
	ioutil.WriteFile(filename, content.Bytes(), 0755)
}
