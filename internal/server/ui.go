package server
import "net/http"
func(s *Server)dashboard(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html; charset=utf-8");w.Write([]byte(dashHTML))}
const dashHTML=`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Headcount</title>
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rl:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--mono:'JetBrains Mono',Consolas,monospace;--serif:'Libre Baskerville',Georgia,serif}*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);font-size:13px;line-height:1.6}.hdr{padding:.6rem 1.2rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-family:var(--serif);font-size:1rem}.hdr h1 span{color:var(--rl)}.main{max-width:800px;margin:0 auto;padding:1rem}.overview{display:flex;gap:1.5rem;margin-bottom:1.2rem;font-size:.7rem;color:var(--leather)}.overview .stat b{display:block;font-size:1.4rem;color:var(--cream)}.section{margin-bottom:1.5rem}.section-title{font-size:.65rem;text-transform:uppercase;letter-spacing:2px;color:var(--rust);margin-bottom:.5rem}.top-row{display:flex;align-items:center;gap:.5rem;padding:.3rem .5rem;border-bottom:1px solid var(--bg3);font-size:.75rem}.top-name{flex:1}.top-bar{height:4px;background:var(--bg3);flex:2;border-radius:2px;overflow:hidden}.top-fill{height:100%;background:var(--rl)}.top-count{width:40px;text-align:right;color:var(--gold)}.evt-row{font-size:.68rem;padding:.2rem .5rem;border-bottom:1px solid var(--bg3);display:flex;gap:.5rem}.empty{text-align:center;padding:2rem;color:var(--cm);font-style:italic;font-family:var(--serif)}</style>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
</head><body><div class="hdr"><h1><span>Headcount</span></h1><span style="font-size:.65rem;color:var(--cm)">POST /api/track</span></div>
<div class="main">
<div id="upgrade-banner" class="upgrade" style="display:none">
  <strong style="color:var(--cream)">Free tier</strong> — 1 site, 1K events/mo. <a href="https://stockyard.dev/headcount/" target="_blank">Upgrade to Pro for $1.99/mo →</a>
</div>
<div class="overview" id="ov"></div>
<div class="section"><div class="section-title">Top Pages (30d)</div><div id="topPages"></div></div>
<div class="section"><div class="section-title">Top Referrers (30d)</div><div id="topRefs"></div></div>
<div class="section"><div class="section-title">Recent Events</div><div id="events"></div></div>
</div>
<script>
async function api(u){return(await fetch(u)).json()}
function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')}
function timeAgo(d){if(!d)return'';const s=Math.floor((Date.now()-new Date(d))/1e3);if(s<60)return s+'s';if(s<3600)return Math.floor(s/60)+'m';return Math.floor(s/3600)+'h'}
function renderTop(items,el){const max=Math.max(1,...(items||[]).map(i=>i.count));el.innerHTML=(items||[]).length?(items||[]).map(i=>'<div class="top-row"><span class="top-name">'+esc(i.name)+'</span><div class="top-bar"><div class="top-fill" style="width:'+Math.round(i.count/max*100)+'%"></div></div><span class="top-count">'+i.count+'</span></div>').join(''):'<div class="empty">No data yet.</div>'}
async function load(){
const[sd,pd,rd,ed]=await Promise.all([api('/api/stats'),api('/api/top/pages'),api('/api/top/referrers'),api('/api/events')]);
document.getElementById('ov').innerHTML='<div class="stat"><b>'+sd.today+'</b>Today</div><div class="stat"><b>'+sd.total_events+'</b>Total Events</div><div class="stat"><b>'+sd.unique_users+'</b>Users (30d)</div><div class="stat"><b>'+sd.unique_sessions+'</b>Sessions (30d)</div>';
renderTop(pd.pages,document.getElementById('topPages'));renderTop(rd.referrers,document.getElementById('topRefs'));
const events=(ed.events||[]).slice(0,20);
document.getElementById('events').innerHTML=events.length?events.map(e=>'<div class="evt-row"><span style="color:var(--gold);width:60px">'+esc(e.name)+'</span><span style="flex:1">'+esc(e.page)+'</span><span style="color:var(--cm)">'+esc(e.user_id||e.ip)+'</span><span style="color:var(--cm)">'+timeAgo(e.created_at)+'</span></div>').join(''):'<div class="empty">No events tracked yet.</div>'}
load();setInterval(load,10000)
fetch('/api/tier').then(r=>r.json()).then(j=>{if(j.tier==='free'){document.getElementById('upgrade-banner').style.display='block';var b=document.getElementById('tier-badge');if(b){b.className='badge badge-free';b.textContent='Free'}}else{var b=document.getElementById('tier-badge');if(b){b.className='badge badge-pro';b.textContent='Pro'}}}).catch(()=>{var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'});
</script></body></html>`
