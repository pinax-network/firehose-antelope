// Copyright 2021 dfuse Platform Inc. //
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codec

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/streamingfast/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logging.TestingOverride()
}

type ObjectReader func() (interface{}, error)

//func TestReadRealBlock(t *testing.T) {
//	fl := strings.NewReader(dmlogblock)
//	cr := testReaderConsoleReader(t, make(chan string, 10000), func() {})
//	cr.ProcessData(fl)
//
//	r, err := cr.Read()
//	assert.NoError(t, err)
//	fmt.Println(r)
//}

func TestParseFromFile(t *testing.T) {
	tests := []struct {
		deepMindFile     string
		expectedPanicErr error
	}{
		{"testdata/deep-mind.dmlog", nil},
	}

	for _, test := range tests {
		t.Run(strings.Replace(test.deepMindFile, "testdata/", "", 1), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					require.Equal(t, test.expectedPanicErr, r)
				}
			}()

			cr := testFileConsoleReader(t, test.deepMindFile)
			buf := &bytes.Buffer{}
			buf.Write([]byte("["))

			for first := true; true; first = false {
				var reader ObjectReader = cr.Read
				out, err := reader()
				if v, ok := out.(proto.Message); ok && !isNil(v) {
					if !first {
						buf.Write([]byte(","))
					}

					// FIXMME: jsonpb needs to be updated to latest version of used gRPC
					//         elements. We are disaligned and using that breaks now.
					//         Needs to check what is the latest way to properly serialize
					//         Proto generated struct to JSON.
					// value, err := jsonpb.MarshalIndentToString(v, "  ")
					// require.NoError(t, err)

					value, err := json.MarshalIndent(v, "", "  ")
					require.NoError(t, err)

					buf.Write([]byte(value))
				}

				if err == io.EOF {
					break
				}

				if len(buf.Bytes()) != 0 {
					buf.Write([]byte("\n"))
				}

				require.NoError(t, err)
			}
			buf.Write([]byte("]"))

			goldenFile := test.deepMindFile + ".golden.json"
			if os.Getenv("GOLDEN_UPDATE") == "true" {
				ioutil.WriteFile(goldenFile, buf.Bytes(), os.ModePerm)
			}

			cnt, err := ioutil.ReadFile(goldenFile)
			require.NoError(t, err)

			if !assert.Equal(t, string(cnt), buf.String()) {
				t.Error("previous diff:\n" + unifiedDiff(t, cnt, buf.Bytes()))
			}
		})
	}
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

func testFileConsoleReader(t *testing.T, filename string) *ConsoleReader {
	t.Helper()

	fl, err := os.Open(filename)
	require.NoError(t, err)

	cr := testReaderConsoleReader(t, make(chan string, 10000), func() { fl.Close() })

	go cr.ProcessData(fl)

	return cr
}

func testReaderConsoleReader(t *testing.T, lines chan string, closer func()) *ConsoleReader {
	t.Helper()

	l := &ConsoleReader{
		lines: lines,
		close: closer,
		ctx: &parseCtx{
			blockMetas: newBlockMetaHeap(blockMetaGetterFunc(func(id string) (*blockMeta, error) {
				return &blockMeta{
					id:        "id.0",
					number:    0,
					blockTime: time.Now(),
				}, nil
			})),
		},
	}

	return l
}

func unifiedDiff(t *testing.T, cnt1, cnt2 []byte) string {
	file1 := "/tmp/gotests-linediff-1"
	file2 := "/tmp/gotests-linediff-2"
	err := ioutil.WriteFile(file1, cnt1, 0600)
	require.NoError(t, err)

	err = ioutil.WriteFile(file2, cnt2, 0600)
	require.NoError(t, err)

	cmd := exec.Command("diff", "-u", file1, file2)
	out, _ := cmd.Output()

	return string(out)
}

var dmlogblock = "DMLOG BLOCK 68081208 eb50f68f4e0ec336860bcc2186b0ac652ac43b356ed5b36ecfd196d7faae4826 0a056e6f64653312f81a08b8acbb201a220a20142574e2cfe679d0d3a4c31cad5d16825701d7e6a1c9223ca8ce800af41eab6c22220a2076a9d927e935feb3edc93004a1e92b4d45a81bfe35c289e65a16262fe0a531072a220a20eb50f68f4e0ec336860bcc2186b0ac652ac43b356ed5b36ecfd196d7faae482632220a207c3e05f6edbb312807ac35925a93d497690750a2a2bab757cb05f09e6ab7f49d3a220a2000a3718d42a4bd29cdbd42a831207859fc02dba3126b546d6cb7764985e6f44c42220a207a4fde36a426deaf5e1252f5718a06e1787f8d12cb7da5a18057fa9ffb66ce0d4a220a200a5048a3a073adde276c77a057cc3af1eeeee096a73defb6400e93091490b38552220a2066687aadf862bd776c8fc18b8e9f8e20089714856ee233b3902a591d0d5f29255a220a2066687aadf862bd776c8fc18b8e9f8e20089714856ee233b3902a591d0d5f29256a220a200000000000000000000000000000000000000000000000000000000000000000709381d7eab5b3ecd716789381d7eab5b3ecd7168201220a205ad7408e42af789b4d8fcfcd5c5cf74e164edfbfe57bfb5303de685838a0e539920101009a01120a1000000000000000000000000005f5e100aa01120a10000069f82b050f5a87bf073deba3383bc201220a206032dcf1c9fcdff51503d2d18d77e1abf2c1fa6dfa1d13dfdac0292373a23583d201220a20cca7e10bb71e633d2d117549da079557342b1a7740c63bce7b4df31ba048eb7cda01220a208558e0438727625285d8d679345696d9c043f0a62ec4716fb72a60f18d4fe09fe201220a20522696708d6af8f5a9ca0bb3c98ac5fbbfe65c49512d6fa16d4c67b5d3780975f20142124049979ec22f85eef7d51ff5fcecfc2f8edef4aff91e5fdf272f0135c3439617e98e7159521113ebb0c1cfb9d627d2e3a7ec1f5f985e0931f183163081f1ed3a06f20142124001e51c9a3b0f26c69050cc8c97bdb1463efa042df31097dc3410b9d310842121adca9303952adf9a8146840ecbe405d976781170d7dc54fee13c03a38530ed05f20142124049979ec22f85eef7d51ff5fcecfc2f8edef4aff91e5fdf272f0135c3439617e98e7159521113ebb0c1cfb9d627d2e3a7ec1f5f985e0931f183163081f1ed3a06f201421240f7e10b38d179268e383486b2a5a21da453c8e4a33667c6b411207cda1c310b5a34130207d24dff308d18ac91639ce873fc60a82cb61710b98d98e454a6949107f20142124049979ec22f85eef7d51ff5fcecfc2f8edef4aff91e5fdf272f0135c3439617e98e7159521113ebb0c1cfb9d627d2e3a7ec1f5f985e0931f183163081f1ed3a06f20142124006e6c6a53e53cc1424b1316f4c6bf61cd82e14684e84661559c0c4104135f457b86dd748dc387da715c274fefff27e5d43917a1fa0a226d78406e05806abcb02f201421240c85ed01cb1d29e427a8b6531f2d622f5f26daa01b11f1528357468143df83fe1e443649c8b98fa318e74a3fc90d2fc2216ee6c3bdb61d273b1f5f76531515504f2014212401a7412b90b12b397637658b9c6f0740f236b178ab4b16e7d74bc4cdd7a27a5d0c1fc78341f90836d1a5a364828d3a8baf2324fe5211a407bf5500fbd88602a02f20142124066cba75e898ea8be0b82ccee3007a810c80d9e5ce198f8d1c2b6132e3c4fd7f16920111806aab0dee4031e10635bbbc9b1f9cb74abecdea4f7b7360b2ea8560ef2014212402691c956f9bb9666d3d97ad8ae72f9d2a17e1afd97bf80963ca8ca3c26b2ed5926ab3036e9ca6c2aa329bdd5d95c40a194fe94d4e3bb7dded248b1b44bcc1e00f20142124049979ec22f85eef7d51ff5fcecfc2f8edef4aff91e5fdf272f0135c3439617e98e7159521113ebb0c1cfb9d627d2e3a7ec1f5f985e0931f183163081f1ed3a06f2014212404b9977c2399cc43d8bc8c615e034b706956d116907ba0cedd5312bb51a75e9f65809d5175843027b430f611c1903eaecb7391a5794c64aae2de5475929633c0cf2014212408946f0efce091654c11809342ce3e4560a3be884a036a22e5adcd79397f790f8ccc704dbaa2bed78ba14ae66e45cff539d07d93ac22cfc6ccf0f7ed237b69607f201421240b0e99ef9b8f7a06fbacfcda2c2d11827e0f6cc6f5432991c2a3e79da8df36aad395b74e46ad2409f67b2bcf539eeb7711702db4c7a727d4bcd8c477c96ebec0ef2014212400dc546bdfefd5a28d74be24817e5d4349599a4eb7635a121d642665c3613bb0676c57be536886815371b85515c2255c19b12f9c6f791943ab9dd239aff8cf703f2014212402868fd8f66ea317cfb9c0a4d17b741076d86155233f7f9de8db0cd621fa2106e8f52d142b2b4cfa62faecf095748601e1770697899e13a9f232a88581c25c409f201421240e3e96da278f02ff78d964481bed368eec64bc1010d9103ecd54768a9da961f845905ec173ee83446b4bf392f5172ad77896450ec407632b421efd83eb33a4c0cf2014212409fc0b82581f59919a23d10be83444e9fc6088c5676eddda2de03c5c004f33f6653ea2ddf555d235aabaf35e97c4d06f43975548f78ebac51c2f6975dd11ae700f201421240838fb5d1f728d669392c50b63de0fabe54e17256967dfd4b2e2cfa76890be218c835c46b3e414f4c7b15a723391d331eabf3284bb335d84ec0ee8ae42fb17208f2014212405b222b9bea6e7211cd209e41205512f7766984404d400c914541c52e9cdde482c91a869f9e76e429cc72f330bd4375e172d2e16f857983dbd9e43a0b3c64a00af201421240a860e728705c8883b1257537af4b2f476cafbc077b16f71e5a643a94f112ccac2222d8f49c2aabfab0865ff1bda7eadc7efbd575a43b25027404accdca1e9c03f2014212404cbb5865834d219e8c716513e81e58ee5b1a0703c77f637887a1e37b2cf65dbdef70612cfc051d4f6590745a4d21dfedfe4bf0ec435211190bda64a456310f02f201421240777b5e1d0ad4847a9b7504e2619a317ee9ba17e7876f0fcb399bb09d8043270a873748edd968f987e22e7514dd772b0858f3f7ecb50b0a517ac3d8b9421e3f0ef2014212402a4f48ef75a465585f94d5b91686a9e94087fd11389f57da0404cf8d54b901dfd61eb003863e007e5d798ab316a7a361e8e56e6a40b08ab6f1a6548139fd250bf20142124035f29dc5cecf1e3637eeb763414376dd90ee371e6f65cbcf868b968cac1c195f0c05af79efa55f2829de4264ea840dda6ba18f4adf9b56792305dde34e269309f201421240d564fd08a4997b92379bdf20ef88a7428591321865fd896f7a0f6829d09a9d2506d603bfad3c65b59a59ddd7aef65b3ffb68063e3e030890fa0b9a34ff3ca807f2014212401d6042d71caa806164d8bce0d3bce21141c5ac62ec0686582d8a48c75ab5aaf246bc4c573e41eca8b0ee14510877790252d02a583178d997f7bad7662b3ec802f201421240861fae0e760b3c0cf792d287c0c7ecff461744e26acaad37f676b135be623b75469d97919c489d93b035d93ebda96181203e222de9a906971c28bb41ae06df00f201421240f01159400f2584ac04d75d29a9fff53b4136c1a8ee1da3ee60ec2a6ca8915743a3c2c59d6ec4b4b1f8dcbc8460ff3b24b9c4b358cc2940f52f79517310367f0ff2014212400057b651b93235cf1e5df0c80126240b8e3323ad2f0957055a4a16cfa8e8c2ab367e5b49f7f003089b8440117f309e4bdc3b5b1f9d1525d0f40bfcada2b1b500f2014212409d732d5ac0545382c83a7c44942488c5949d06f2672f993dcfaa07bb5fc4939fe64ea2d9818dc8b5acfc3ef2a9a63e3169fc7c410503f47a95194a7f5b424001f201421240a790e21d3dd2d246e2b4caa55ed98cb126280a755a0783fca4ddff693dbc56a57eb81ffdf2e726c158d84539ba2fb10bca5a06ae4484206b942c458173a9640cf201421240c1d665ccad5a83c5dd4e14a24f06c250c6e5e367ca6196d1b9c6568889f6112e9170e4e37ab9279eda6cd7720b9ae49bb82f36e899016b9d623e7aa1b36f330bf201421240ae70583c3f7d592ec690ed0970eea446b59f10957c4ac07a12e9e9c4e0cddaa50cf81ff130654f7366d1c7a6507fc66cf769ad07fd0953ec02d53b82aa23330cf201421240c09fd57868e26ef013bf28d71aa74428c3bd68efae9976e967eea0772749647ebaef193e0d954d832a6602176c05fd503018e072d44ddc027ea1ab6f688c0c0af201421240945b57d581bc231e8492d3617c469022813ebbc0a63b460c6edf46bc51ad9209628a6cbd404a2ec4de300b494c1e2aad623e55b08a5d1904b8a9bf26c246b90af2014212404d32a96a46d97f9020a945c8229329eaed49f49b2faf4329f9a802373a1c7a83b6573cc44a11343a4372894aaf756468409b21e44c87e1ae2f497d895c827309f201421240e667910d5453bb4514ab93657f0a4fc761e0dc8bc84758933b9ef9d3ae89d5ab07442170a00689d8a2205f207acac18be83825112904f40506d62d37ed15c108f20142124041c09d5a9a1838607a706152019e2b52cac6ebef71825c1e79f86f106ca9aff58a5dbf4a90ad6926bed52cd860c4c1efebf989774964b3575014c4523c024509f2014212402b36386484cd45d18dc3e0f902fa5e06c963815637322c6a3b12130f84eea759beaacd4324941ea82373b302e64bfd16a337a8105d89c839aa3fe5d774b5760ffa0142124018ad1747aa2a604cdee3a1bb3dc66ee01dcce6218d181fcb3a7dc4689430541bf664cd15b4e3138557d80c192153ef10df0bceefb92bed03b3973a2734ffe00c8002301af0020a205ca9b9cdf50ffddf82a41794a02bb0b2326ce175c04729fe5499b0965f586b5d1220cca7e10bb71e633d2d117549da079557342b1a7740c63bce7b4df31ba048eb7c1a200000000000000000000000000000000000000000000000000000000000000000222004e5ba48500ac1b67057befedd47aab4c7d32ae6f0166e612dbe591c1528740b2a204256a8d33a1b65ca1f0402cef0d3e56810cb67961f57a02a9039631cf3af34e9300838b6acbb2040b6acbb205880809aa6eaafe30162120a10000000000000000000000000000000006a120a10000000000000000000000000000000007220ee9b92e324e0341a965daf29b39555030e2c0e5591d7d3e2bd05ff785f946b057a2000000000000000000000000000000000000000000000000000000000000000008a01421240db8273eca79bfefc65e655de4e46639097ae2545b79b9cecbe2c014af585a497c89031c608d89bbacef897b6a11329b35f64abd50e536ce5ca5829818db52d062200"
