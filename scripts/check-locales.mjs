import { readFileSync, readdirSync } from 'node:fs';
for (const directory of ['../web/src/locales/', '../demo/src/locales/']) {
const root = new URL(directory, import.meta.url);
const base = JSON.parse(readFileSync(new URL('en.json',root),'utf8'));
for(const file of readdirSync(root).filter(f=>f.endsWith('.json'))){
 const values=JSON.parse(readFileSync(new URL(file,root),'utf8'));
 if(JSON.stringify(Object.keys(base).sort())!==JSON.stringify(Object.keys(values).sort()))throw Error(`Locale keys differ: ${file}`);
 for(const [key,value] of Object.entries(values))if(typeof value!=='string'||!value.trim())throw Error(`Empty translation: ${file}/${key}`);
}
}
console.log('Locale keys and values passed');
