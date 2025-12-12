// Wait until the DOM is ready before grabbing elements
let translationAmnt = 10;

document.addEventListener("DOMContentLoaded", () => {
    init();
    setTimeout(() => {
        animate();
    }, 2500)
});

async function getServiceData() {
    const res = await fetch("/api");

    if (!res.ok) {
        alert("Server error!");
        return [];
    }

    const data = await res.json();
    return data;
}

async function init() {
    const cards = document.getElementById("cards");

    if (!cards) {
        console.error("No element with id='cards' found");
        return;
    }

    try {
        const jsonData = await getServiceData();

        if (!jsonData) {
            cards.innerHTML = "<div class=\"card\"><p>No Service Data to Display</p></div>"
        }

        if (Array.isArray(jsonData)) {
            jsonData.forEach(svc => {
                cards.innerHTML += `
                    <div class="card">
                        <h1 class="svcName">Server URL: <a href="${svc.name}">${svc.name}</a></h1>
                        <p>Status: ${svc.response === "200" ? "✅" : "❌"} in ${svc.response_time}</p>
                        <p class="svcHttpRes">HTTP Response: ${svc.response}</p>
                        <h2>API Response</h2>
                        <div class="svcData">${svc.data ? svc.data : "No API setup in configuration"}</div>
                    </div>
                `;
            });
        } else {
            console.error("Expected array from /api, got:", jsonData);
        }
    } catch (err) {
        console.error("Error fetching service data:", err);
    }

    await getSchedule();
    await getUptimeAverages()
    
}

function animate() {
    translationAmnt += 1;    
    document.getElementById("logo").style.transform =
        `translate(${translationAmnt}px, -${translationAmnt}px)`;
    if (translationAmnt < 500) {
        requestAnimationFrame(animate);
    } else {
        document.getElementById("logo").style.display = "none";
    }
    
}
// eslint-disable-next-line no-unused-vars
async function updateSchedule() {
    /**
     * @type {string}
     */
    const ts = document.getElementById("timespan").value;
    /**
     * @type {string}
     */
    const ti = document.getElementById("timeInterval").value;

    const req = await fetch("/api/schedule", {
        method: "POST",
        body: JSON.stringify({
            timespan: Number(ts),
            timeInterval: ti,
        }),
    });

    if (!req.ok) {
        alert("Update failed!");
    }
    const timeCalc = () => {
        switch (ti.charAt(0).toLowerCase()) {
            case 's':
                return "seconds"
            case 'm':
                return "minutes"
            case 'h':
                return "hours"
            default:
                return "Undefined"
        }
    }
    document.getElementById("currentSchedule").innerHTML = `${ts} ${timeCalc()}`;
}

async function getSchedule() {
    try {
        const res = await fetch("/api/schedule");

        if (!res.ok) {
            alert("Server error!");
            async () => {
                const err = await res.text();

                throw new Error(err)
            }
        }

        const jsonData = await res.json();

        const s = document.getElementById("schedule");

        s.innerHTML = `
        <h3>Current Settings</h3>
        <p id="currentSchedule">${jsonData.timespan} ${jsonData.timeInterval}</p>
        <form id="scheduleForm">
            <label>Timespan</label>
            <p id="sliderValue">Slider Value: </p>
            <input id="timespan" type="range" min="1" max="60" default="30" value="${jsonData.timespan}" required>
            <label>Interval (sec, min, hr)</label>
            <input id="timeInterval" name="interval" type="text" required>
        </form>
        <button onclick="updateSchedule()">Update</button>
        `;
    } catch (err) {
        console.error("Error fetching service data:", err);
    }

    const timespanEl = document.getElementById("timespan");

    timespanEl.addEventListener('input', () => {
        document.getElementById("sliderValue").innerText = `Slider Value: ${timespanEl.value}`;
    })
}

async function getUptimeAverages() {
    const ctx = document.getElementById('averageChart')

    const res = await fetch("/api/uptime", {
        method: "GET"
    });

    if (!res.ok) {
        console.log(`Error fetching data: ${res.statusText}`)
    }

    const json = await res.json()
    
    let xLabels = []
    let yLabels = []

    for (let i = 0; i < json.length; i++) {
        xLabels.push(json[i].name)
        yLabels.push(json[i].average)
    }
    /**
     * Chart is imported from https://cdn.jsdelivr.net/npm/chart.js CDN
     * Used to display average information in chart form
     */
    /* global Chart */
    new Chart(ctx, {
        type: 'bar',
        data: {
        labels: xLabels, // X-axis labels
        datasets: [{
            label: 'Total Uptime Average',
            data: yLabels, // Y-axis data
            backgroundColor: [
            'red',
            'blue',
            'black',
            'white'
            ],
            borderWidth: 1
        }]
        },
        options: {
        scales: {
        y: { beginAtZero: true }
        }
    }
    }
    );

}