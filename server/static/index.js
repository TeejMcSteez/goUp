// Wait until the DOM is ready before grabbing elements
document.addEventListener("DOMContentLoaded", () => {
    init();
});

async function getServiceData() {
    const res = await fetch("http://localhost:8080/api");

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

        if (Array.isArray(jsonData)) {
            jsonData.forEach((svc, idx) => {
                cards.innerHTML += `
                    <div class="card">
                        <h1 class="svcName">Server URL: <a href="${svc.name}">${svc.name}</a></h1>
                        <p>Status: ${svc.response === "200" ? "✅" : "❌"}</p>
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
}
