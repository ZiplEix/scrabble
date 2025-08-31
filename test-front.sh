set -e;

cd frontend;

npm run dev >/dev/null 2>&1 & echo $$! > ../.front.pid;

cd ..;

echo "Attente du front sur :3000...";
for i in $(seq 1 60); do
    if curl -sf http://localhost:3000/ >/dev/null; then
        echo "Frontend prêt"; break;
    fi;
    sleep 0.5;
done;

cd frontend && npx cypress run;

cd ..;
if [ -f .front.pid ]; then
    kill $(cat .front.pid) 2>/dev/null || true;
    rm -f .front.pid;
fi
