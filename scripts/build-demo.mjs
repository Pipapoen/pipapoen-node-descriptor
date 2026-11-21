// Build the production Go codec for the browser with its matching Go runtime.
import {execFileSync} from 'node:child_process';
import {readFileSync,writeFileSync,mkdirSync,rmSync,copyFileSync} from 'node:fs';
import {fileURLToPath} from 'node:url';
import {join} from 'node:path';
import {createHash} from 'node:crypto';
import {gzipSync} from 'node:zlib';

const root=fileURLToPath(new URL('../',import.meta.url));
const build=join(root,'build','demo');
const output=join(root,'demo','public','runtime');
mkdirSync(build,{recursive:true});rmSync(output,{recursive:true,force:true});mkdirSync(output,{recursive:true});
execFileSync('go',['build','-trimpath','-ldflags=-s -w','-o',join(build,'pnd.wasm'),'./cmd/demo-wasm'],{
 cwd:root,stdio:'inherit',env:{...process.env,GOOS:'js',GOARCH:'wasm',CGO_ENABLED:'0'}
});
const goroot=execFileSync('go',['env','GOROOT'],{cwd:root,encoding:'utf8'}).trim();
const wasm=gzipSync(readFileSync(join(build,'pnd.wasm')),{level:9});
const js=readFileSync(join(goroot,'lib','wasm','wasm_exec.js'));
const hash=value=>createHash('sha256').update(value).digest('hex').slice(0,16);
const module=`pnd-${hash(wasm)}.wasm.gz`,runtime=`wasm_exec-${hash(js)}.js`;
writeFileSync(join(output,module),wasm);writeFileSync(join(output,runtime),js);
writeFileSync(join(output,'manifest.json'),JSON.stringify({module,runtime}));
const licenses=['go_LICENSE','github.com_joshuajorel_scrambler_LICENSE','basecoat_LICENSE','tailwindcss_LICENSE','geonames_NOTICE.txt'];
for(const name of licenses)copyFileSync(join(root,'third_party','licenses',name),join(output,name));
copyFileSync(join(root,'LICENSE'),join(output,'PND_LICENSE'));
writeFileSync(join(output,'NOTICE.txt'),'PND browser demo uses the same interface, codec and static geographic data as the deployed application.\n\nPublic demonstration key: 32 zero bytes.\n\nCopyright and license notices in this directory:\nPND_LICENSE\n'+licenses.join('\n')+'\n');
console.log(`Browser codec and geography: ${(wasm.length/1024/1024).toFixed(2)} MiB compressed.`);
