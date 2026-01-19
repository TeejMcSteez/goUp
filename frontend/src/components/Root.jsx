import Header from "./Header.jsx";
import Update from "./Update.jsx";
import Services from "./Services.jsx";
import UptimeChart from "./UptimeChart.jsx";
import QuickServices from "./QuickServices.jsx";
import DatabaseInfo from "./DatabaseInfo.jsx";
import ErrorLogViewer from "./ErrorLogViewer.jsx";

export default function Root() {
    return (
        <div id="container">
            <Header />
            <QuickServices />
            <Update />
            <DatabaseInfo />
            <Services />
            <UptimeChart />
            <ErrorLogViewer />
        </div>
    )
}