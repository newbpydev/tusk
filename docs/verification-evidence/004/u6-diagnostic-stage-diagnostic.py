import json,os,sqlite3,subprocess,tempfile,time,pathlib
with tempfile.TemporaryDirectory(prefix='tusk-stage-') as d:
 env={'PATH':os.environ['PATH'],'HOME':d,'TUSK_DB_PATH':d+'/db','TUSK_TIMEZONE':'UTC','TERM':'dumb'}
 subprocess.run(['/tmp/tusk-timing-bin','list','--all'],env=env,check=True,capture_output=True)
 con=sqlite3.connect(d+'/db')
 now='2026-09-28T12:00:00.000000000Z'
 for i in range(1000):
  status=['todo','in-progress','blocked','done'][i%4]
  if i<9: status='todo'
  parent=None if i%10==0 else f'fixture-{i-1 if i<10 else i//10*10:06d}'
  con.execute('INSERT INTO tasks (id,title,description,status,priority,progress,parent_id,tags,due_date,created_at,updated_at,completed_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)',(f'fixture-{i:06d}','t'*64,'n'*128,status,i%4+1,100 if status=='done' else 0,parent,'["alpha","beta","gamma"]',None if i%3==0 else now,now,now,now if status=='done' else None))
 con.commit();con.close()
 samples=[]
 for i in range(105):
  start=time.perf_counter_ns()
  p=subprocess.run(['/tmp/tusk-timing-bin','list','--all','--json'],env=env,capture_output=True,check=True)
  wall=time.perf_counter_ns()-start
  stage=json.loads(p.stderr);stage.update(index=i,warmup=i<5,wall=wall,bytes=len(p.stdout))
  assert len(json.loads(p.stdout))==1000
  samples.append(stage)
 pathlib.Path('/tmp/tusk-004-u6-stage-diagnostic.json').write_text(json.dumps(samples,indent=2))
 for s in sorted(samples[5:],key=lambda s:s['wall'])[-10:]: print({k:round(v/1e6,3) if k not in ('index','warmup','bytes') else v for k,v in s.items()})
