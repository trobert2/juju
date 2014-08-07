// Copyright 2013 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

// The format tests are white box tests, meaning that the tests are in the
// same package as the code, as all the format details are internal to the
// package.

package agent

import (
	"io/ioutil"
	"path/filepath"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	gc "launchpad.net/gocheck"

	"github.com/juju/juju/testing"
	"github.com/juju/juju/version"
)

type format_1_16Suite struct {
	testing.BaseSuite
}

var _ = gc.Suite(&format_1_16Suite{})

func (*format_1_16Suite) TestStatePortParsed(c *gc.C) {
	dataDir := c.MkDir()
	formatPath := filepath.Join(dataDir, legacyFormatFilename)
	err := utils.AtomicWriteFile(formatPath, []byte(legacyFormatFileContents), 0600)
	c.Assert(err, gc.IsNil)
	configPath := filepath.Join(dataDir, agentConfigFilename)
	err = utils.AtomicWriteFile(configPath, []byte(stateMachineConfigData), 0600)
	c.Assert(err, gc.IsNil)
	readConfig, err := ReadConfig(configPath)
	c.Assert(err, gc.IsNil)
	info, available := readConfig.StateServingInfo()
	c.Assert(available, gc.Equals, true)
	c.Assert(info.StatePort, gc.Equals, 37017)
}

func (*format_1_16Suite) TestReadConfReadsLegacyFormatAndWritesNew(c *gc.C) {
	dataDir := c.MkDir()
	formatPath := filepath.Join(dataDir, legacyFormatFilename)
	err := utils.AtomicWriteFile(formatPath, []byte(legacyFormatFileContents), 0600)
	c.Assert(err, gc.IsNil)
	configPath := filepath.Join(dataDir, agentConfigFilename)
	err = utils.AtomicWriteFile(configPath, []byte(agentConfig1_16Contents), 0600)
	c.Assert(err, gc.IsNil)

	config, err := ReadConfig(configPath)
	c.Assert(err, gc.IsNil)
	c.Assert(config, gc.NotNil)
	// Test we wrote a currently valid config.
	config, err = ReadConfig(configPath)
	c.Assert(err, gc.IsNil)
	c.Assert(config, gc.NotNil)
	c.Assert(config.UpgradedToVersion(), jc.DeepEquals, version.MustParse("1.16.0"))
	c.Assert(config.Jobs(), gc.HasLen, 0)

	// Old format was deleted.
	assertFileNotExist(c, formatPath)
	// And new contents were written.
	data, err := ioutil.ReadFile(configPath)
	c.Assert(err, gc.IsNil)
	c.Assert(string(data), gc.Not(gc.Equals), agentConfig1_16Contents)
}

const legacyFormatFileContents = "format 1.16"

var agentConfig1_16Contents = `
tag: machine-0
nonce: user-admin:bootstrap
cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNXekNDQWNhZ0F3SUJBZ0lCQURBTEJna3Foa2lHOXcwQkFRVXdRekVOTUFzR0ExVUVDaE1FYW5WcWRURXkKTURBR0ExVUVBd3dwYW5WcWRTMW5aVzVsY21GMFpXUWdRMEVnWm05eUlHVnVkbWx5YjI1dFpXNTBJQ0pzYjJOaApiQ0l3SGhjTk1UUXdNekExTVRJeE9ERTJXaGNOTWpRd016QTFNVEl5TXpFMldqQkRNUTB3Q3dZRFZRUUtFd1JxCmRXcDFNVEl3TUFZRFZRUUREQ2xxZFdwMUxXZGxibVZ5WVhSbFpDQkRRU0JtYjNJZ1pXNTJhWEp2Ym0xbGJuUWcKSW14dlkyRnNJakNCbnpBTkJna3Foa2lHOXcwQkFRRUZBQU9CalFBd2dZa0NnWUVBd3NaVUg3NUZGSW1QUWVGSgpaVnVYcmlUWmNYdlNQMnk0VDJaSU5WNlVrY2E5VFdXb01XaWlPYm4yNk03MjNGQllPczh3WHRYNEUxZ2l1amxYCmZGeHNFckloczEyVXQ1S3JOVkkyMlEydCtVOGViakZMUHJiUE5Fb3pzdnU3UzFjZklFbjBXTVg4MWRBaENOMnQKVkxGaC9hS3NqSHdDLzJ5Y3Z0VSttTngyVG5FQ0F3RUFBYU5qTUdFd0RnWURWUjBQQVFIL0JBUURBZ0NrTUE4RwpBMVVkRXdFQi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZKVUxKZVlIbERsdlJ3T0owcWdyemcwclZGZUVNQjhHCkExVWRJd1FZTUJhQUZKVUxKZVlIbERsdlJ3T0owcWdyemcwclZGZUVNQXNHQ1NxR1NJYjNEUUVCQlFPQmdRQ2UKRlRZbThsWkVYZUp1cEdPc3pwc2pjaHNSMEFxeXROZ1dxQmE1cWUyMS9xS2R3TUFSQkNFMTU3eUxGVnl6MVoycQp2YVhVNy9VKzdKbGNiWmtHRHJ5djE2S2UwK2RIY3NEdG5jR2FOVkZKMTAxYnNJNG1sVEkzQWpQNDErNG5mQ0VlCmhwalRvYm1YdlBhOFN1NGhQYTBFc1E4bXFaZGFabmdwRU0vb1JiZ0RMdz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
stateaddresses:
- localhost:37017
statepassword: OlUMkte5J3Ss0CH9yxedilIC
apiaddresses:
- localhost:17070
apipassword: OlUMkte5J3Ss0CH9yxedilIC
oldpassword: oBlMbFUGvCb2PMFgYVzjS6GD
values:
  PROVIDER_TYPE: local
  SHARED_STORAGE_ADDR: 10.0.3.1:8041
  SHARED_STORAGE_DIR: /home/user/.juju/local/shared-storage
  STORAGE_ADDR: 10.0.3.1:8040
  STORAGE_DIR: /home/user/.juju/local/storage
stateservercert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNJakNDQVkyZ0F3SUJBZ0lCQURBTEJna3Foa2lHOXcwQkFRVXdRekVOTUFzR0ExVUVDaE1FYW5WcWRURXkKTURBR0ExVUVBd3dwYW5WcWRTMW5aVzVsY21GMFpXUWdRMEVnWm05eUlHVnVkbWx5YjI1dFpXNTBJQ0pzYjJOaApiQ0l3SGhjTk1UUXdNekExTVRJeE9ESXlXaGNOTWpRd016QTFNVEl5TXpJeVdqQWJNUTB3Q3dZRFZRUUtFd1JxCmRXcDFNUW93Q0FZRFZRUURFd0VxTUlHZk1BMEdDU3FHU0liM0RRRUJBUVVBQTRHTkFEQ0JpUUtCZ1FDdVA0dTAKQjZtbGs0V0g3SHFvOXhkSFp4TWtCUVRqV2VLTkhERzFMb21SWmc2RHA4Z0VQK0ZNVm5IaUprZW1pQnJNSEk3OAo5bG4zSVRBT0NJT0xna0NkN3ZsaDJub2FheTlSeXpUaG9PZ0RMSzVpR0VidmZDeEFWZThhWDQvbThhOGNLWE9TCmJJZTZFNnVtb0wza0JNaEdiL1QrYW1xbHRjaHVNRXJhanJSVit3SURBUUFCbzFJd1VEQU9CZ05WSFE4QkFmOEUKQkFNQ0FCQXdIUVlEVlIwT0JCWUVGRTV1RFg3UlRjckF2ajFNcWpiU2w1M21pR0NITUI4R0ExVWRJd1FZTUJhQQpGSlVMSmVZSGxEbHZSd09KMHFncnpnMHJWRmVFTUFzR0NTcUdTSWIzRFFFQkJRT0JnUUJUNC8vZkpESUcxM2dxClBiamNnUTN6eHh6TG12STY5Ty8zMFFDbmIrUGZObDRET0U1SktwVE5OTjhkOEJEQWZPYStvWE5neEM3VTZXdjUKZjBYNzEyRnlNdUc3VXJEVkNDY0kxS3JSQ0F0THlPWUREL0ZPblBwSWdVQjF1bFRnOGlRUzdlTjM2d0NEL21wVApsUVVUS2FuU00yMnhnWWJKazlRY1dBSzQ0ZjA4SEE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
stateserverkey: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDV3dJQkFBS0JnUUN1UDR1MEI2bWxrNFdIN0hxbzl4ZEhaeE1rQlFUaldlS05IREcxTG9tUlpnNkRwOGdFClArRk1WbkhpSmtlbWlCck1ISTc4OWxuM0lUQU9DSU9MZ2tDZDd2bGgybm9hYXk5Unl6VGhvT2dETEs1aUdFYnYKZkN4QVZlOGFYNC9tOGE4Y0tYT1NiSWU2RTZ1bW9MM2tCTWhHYi9UK2FtcWx0Y2h1TUVyYWpyUlYrd0lEQVFBQgpBb0dBRERJZ2FoSmJPbDZQNndxUEwwSlVHOGhJRzY1S1FFdHJRdXNsUTRRbFZzcm8yeWdrSkwvLzJlTDNCNWdjClRiaWEvNHhFS2Nwb1U1YThFVTloUGFONU9EYnlkVEsxQ1I3R2JXSGkwWm1LbGZCUlR4bUpxakdKVU1CSmI4a0QKNStpMzlvcXdQS3dnaXoyTVR5SHZKZFFJVHB0ZDVrbEQyYjU1by9YWFRCTnk2NGtDUVFEbXRFWHNTL2kxTm5pSwozZVJkeHM4UVFGN0pKVG5SR042ZUh6ZHlXb242Zjl2ZkxrSDROWUdxcFUydjVBNUl1Nno3K3NJdXVHU2ZSeEI1CktrZVFXdlVQQWtFQXdWcVdlczdmc3NLbUFCZGxER3ozYzNxMjI2eVVaUE00R3lTb1cxYXZsYzJ1VDVYRm9vVUsKNjRpUjJuU2I1OHZ2bGY1RjRRMnJuRjh2cFRLcFJwK0lWUUpBTlcwZ0dFWEx0ZU9FYk54UUMydUQva1o1N09rRApCNnBUdTVpTkZaMWtBSy9sY2p6YktDanorMW5Hc09vR2FNK1ZrdEVTY1JGZ3RBWVlDWWRDQldzYS93SkFROWJXCnlVdmdMTVlpbkJHWlFKelN6VStXN01oR1lJejllSGlLSVZIdTFTNlBKQmsyZUdrWmhiNHEvbXkvYnJxYzJ4R1YKenZxTzVaUjRFUXdQWEZvSTZRSkFkeVdDMllOTTF2a1BuWnJqZzNQQXFHRHJQMHJwNEZ0bFV4alh0ay8vcW9hNgpRcXVYcE9vNjd4THRieW1PTlJTdDFiZGE5ZE5tbGljMFVNZ0JQRHgrYnc9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
apiport: 17070
`[1:]

const configDataWithoutNewAttributes = `
tag: user-omg
nonce: a nonce
cacert: Y2EgY2VydA==
stateaddresses:
- localhost:1234
apiaddresses:
- localhost:1235
oldpassword: sekrit
values: {}
`

const stateMachineConfigData = `
tag: machine-0
nonce: user-admin:bootstrap
cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNXekNDQWNhZ0F3SUJBZ0lCQURBTEJna3Foa2lHOXcwQkFRVXdRekVOTUFzR0ExVUVDaE1FYW5WcWRURXkKTURBR0ExVUVBd3dwYW5WcWRTMW5aVzVsY21GMFpXUWdRMEVnWm05eUlHVnVkbWx5YjI1dFpXNTBJQ0pzYjJOaApiQ0l3SGhjTk1UTXdPVEkzTURZME1ERTFXaGNOTWpNd09USTNNRFkwTlRFMVdqQkRNUTB3Q3dZRFZRUUtFd1JxCmRXcDFNVEl3TUFZRFZRUUREQ2xxZFdwMUxXZGxibVZ5WVhSbFpDQkRRU0JtYjNJZ1pXNTJhWEp2Ym0xbGJuUWcKSW14dlkyRnNJakNCbnpBTkJna3Foa2lHOXcwQkFRRUZBQU9CalFBd2dZa0NnWUVBcWRhYWFVWE9YTFNtcTdhVApKUTNzckFIb3dFUjJnTFcyd1g5dHptMGdqVkZEVVBkdjNQQ3N1b1R6THdkaXhaQ2dJMFpMaGY5cWllYkZkSmpZCjAxOHUrVHovTkJuMzJLdDYzZWM3YmtRWnR3T09jSEZOWDhHZUdRRkVGOVVJcjYzeGxhUnNaMnJybTFlZCszZTgKdDdwendHY2YvdlB0ZmxldlJXRUpIT1l6MVZVQ0F3RUFBYU5qTUdFd0RnWURWUjBQQVFIL0JBUURBZ0NrTUE4RwpBMVVkRXdFQi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZDQlhnaXFpSkhBWVZ5RlA3R1hSS3NkcVlEVzhNQjhHCkExVWRJd1FZTUJhQUZDQlhnaXFpSkhBWVZ5RlA3R1hSS3NkcVlEVzhNQXNHQ1NxR1NJYjNEUUVCQlFPQmdRQWgKTy9JcWRjYnhsNzBpcUMzcHVqNGswbnV6ZFNoOXFlTzZVVktaYkVIWmtLV2J1ejVHK2tBdldaQ0QwcVhjb0JFcgpLc2dKZlNLdDVKWXZUQW1uUnF2dEdLVWN6SGN0WHMyQVBkWWcrRnkvdGd2THFSNGdaeXN4NWs3cVV1MVNITWZhCk5CUlo4YkdBbGZsOXF2Rlo5TkR4NElKUnQzUGh3S1FRWlpmcTkzQm5SQT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
stateaddresses:
- localhost:37017
statepassword: +PCNyLFNAg2f5SN3ig6uHHum
apiaddresses:
- localhost:17070
apipassword: +PCNyLFNAg2f5SN3ig6uHHum
oldpassword: Jc1GMZX/d35BgbQ6F9nxrTY4
values:
  PROVIDER_TYPE: local
  SHARED_STORAGE_ADDR: 10.0.3.1:8041
  SHARED_STORAGE_DIR: /home/rog/.juju/local/shared-storage
  STORAGE_ADDR: 10.0.3.1:8040
  STORAGE_DIR: /home/rog/.juju/local/storage
stateservercert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNJakNDQVkyZ0F3SUJBZ0lCQURBTEJna3Foa2lHOXcwQkFRVXdRekVOTUFzR0ExVUVDaE1FYW5WcWRURXkKTURBR0ExVUVBd3dwYW5WcWRTMW5aVzVsY21GMFpXUWdRMEVnWm05eUlHVnVkbWx5YjI1dFpXNTBJQ0pzYjJOaApiQ0l3SGhjTk1UUXdOREF4TVRNd016SXlXaGNOTWpRd05EQXhNVE13T0RJeVdqQWJNUTB3Q3dZRFZRUUtFd1JxCmRXcDFNUW93Q0FZRFZRUURFd0VxTUlHZk1BMEdDU3FHU0liM0RRRUJBUVVBQTRHTkFEQ0JpUUtCZ1FDa1E1RzEKbUFuQU0wb3REVzVwREo3R3pQbTg5OUtySlVlR0NIZytGV2l0d1RETnJiK0NhYk1TYWRsc3JYb0crYjdETDFIcApXNTdnQXZoNjBTeUFLWHJCVW9tMG1pdVI1QkhYeitpWkZsZDZHS0UySTFIMUlON0pldUdmTURyVUN4WlVYNkdkCjVlcStUU3JvQ3ZPVGxDYWFtNDRkaHd0S1JHMlFQQ2RYbTNSbWxRSURBUUFCbzFJd1VEQU9CZ05WSFE4QkFmOEUKQkFNQ0FCQXdIUVlEVlIwT0JCWUVGTElWeDdmUVJFUkRGZ3hCcWh4b3puMHZueUlXTUI4R0ExVWRJd1FZTUJhQQpGQ0JYZ2lxaUpIQVlWeUZQN0dYUktzZHFZRFc4TUFzR0NTcUdTSWIzRFFFQkJRT0JnUUFKeW9yaEtLa20ySEFBCmNtS2RyRFNyRlZraElxUFlnc0p6STVTOXRBb0lxRDYwMUZ2eVh1aE50STlwR21ZS2tEd1J0Q2JXNy9nL1RMYVIKbVhXcEpqSDRMNlNLbEFkRFFVMVpPejMwRTdlR3F6aXp3dUdTUHB1VDdjUm5wOVVYdEwrRGZPc2N4WDNwNXMvMwpobmJGdFZGVWllejJRVDNoemo4VTRocXlWTENNZkE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
stateserverkey: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWEFJQkFBS0JnUUNrUTVHMW1BbkFNMG90RFc1cERKN0d6UG04OTlLckpVZUdDSGcrRldpdHdURE5yYitDCmFiTVNhZGxzclhvRytiN0RMMUhwVzU3Z0F2aDYwU3lBS1hyQlVvbTBtaXVSNUJIWHoraVpGbGQ2R0tFMkkxSDEKSU43SmV1R2ZNRHJVQ3haVVg2R2Q1ZXErVFNyb0N2T1RsQ2FhbTQ0ZGh3dEtSRzJRUENkWG0zUm1sUUlEQVFBQgpBb0dBY3ZvODFxQTZTd2RicDFkY2JqbUFOZVVwOWNSOStIL2FwWTN1Skg2MXk5R0xTSnlTalVWUks5VmRkRDJsClNaYXNtVkRaQS8rMm9GUlQreHZKQzFoOWJBNm51NzBxczZXUXBQczQ5WGxhSFdNWXJ0dEV5UDVXeVE4ZWNPWlkKazJZeWJsN3ZQVnhOS1VXdk85L0N3MDgyU2FWZUJGbktvSkRxM1NZZHAzYnhWOEVDUVFEQW85cnBibTFnaENkcApIUFNIQU1SY0lOZUpLcHoxM3QxS3ErN1E3YUZOOWsxYnZvWm8yV2FZZ1pRbXBRL1RoNnl3dy9teWJscmxpMUxGCm5Vc25HZzV4QWtFQTJrcDNnV1B1aXN1bHoyMU1hQmtaN0pLUzVKUXkyQlFUM2ZuM0Nua3hFa2xRdGZ2VnFBN1cKMndPbG9acUFBM2ZCRVUycWEyVmptejg1WGZKUVZYbjBaUUpCQUljaTZ2Q1NESnlHV0hjK1hyTk44SEdJZ0dxeQp3QVVpNEMzL3lybzUyTXd1R2pwZnZ6NVNNOHlNS2ZlcUZ4NFdzU2dYY2xTZllaaGhVaUZhcEZ1N3hhRUNRR2lWCmV2SWtGYnFyM1RJbk5JOC9UM3RYc2tjUGRkaXVyZUlSQzdvWjNGZmRobXphVGtBcGMra1VzenRjMFc1WDVzbEsKZzViV3ljVXNvbWlQV3N2SkZUMENRQTRseEVjN0ZKd0xmRTVRMGpoUkc0d0Jjdll5YUtNRzNiQi9YYzlzZU1uUwpjU3RqM2ZzZkIwYTNldENMZW1PTnpaWkV1YVlFUjZiblR6R3BqdFhwQ3lBPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
apiport: 17070
`
