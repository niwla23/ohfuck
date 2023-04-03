import React, { useState, useEffect, useRef } from "react"
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome"
import { faHeartBroken, faCircle } from "@fortawesome/free-solid-svg-icons"

interface MonitorData {
  name: string
  friendlyName: string
  up: boolean
  reason: string
}


const getReason = (providedReason: string, up: boolean) => {
  if (providedReason.length > 0) {
    return providedReason
  } else if (up) {
    return "Everything is fine."
  } else {
    return "Crying. Please fix."
  }
}

const MonitorCard = (props: { monitorData: MonitorData }) => {
  const extraClasses = props.monitorData.up ? "bg-green-700" : "bg-red-700"

  let iconComponent = props.monitorData.up ? (
    <div className="p-2">
      <div className="relative w-4 h-4">
        <FontAwesomeIcon icon={props.monitorData.up ? faCircle : faHeartBroken} size="lg" className="animate-ping absolute inset-0" />
        <FontAwesomeIcon icon={props.monitorData.up ? faCircle : faHeartBroken} size="lg" className="absolute inset-0" />
      </div>
    </div>
  ) : (
    <FontAwesomeIcon icon={faHeartBroken} size="2x" className="" />
  )

  return (
    <div className={`p-4 rounded-md flex gap-4 items-center md:min-h-[100px] ${extraClasses}`}>
      {iconComponent}
      <div>
        <h2 className="font-bold">{props.monitorData.friendlyName}</h2>
        <p>{getReason(props.monitorData.reason, props.monitorData.up)}</p>
      </div>
    </div>
  )
}

const usePrevious = <T extends unknown>(value: T): T | undefined => {
  const ref = useRef<T>()
  useEffect(() => {
    ref.current = value
  })
  return ref.current
}

export default function Home() {
  const [monitors, setMonitors] = useState<MonitorData[]>([])
  const previousMonitors = usePrevious(monitors)
  let intervalId = 0

  const updateMonitors = async () => {
    let resp = await fetch("/api/monitors")
    setMonitors(await resp.json())
  }

  useEffect(() => {
    Notification.requestPermission()

    updateMonitors()
    intervalId = setInterval(updateMonitors, 5000)

    return () => {
      clearInterval(intervalId)
    }
  }, [])

  useEffect(() => {
    for (let monitor of monitors) {
      let oldMonitor = previousMonitors?.find((i) => i.name === monitor.name)
      if (!oldMonitor) {
        return
      }
      if (!monitor.up && oldMonitor?.up) {
        new Notification(`DOWN: ${monitor.friendlyName}`, {
          body: `"${monitor.friendlyName}" is crying. please fix.`,
          icon: "/logo.svg",
        })
      } else if (monitor.up && !oldMonitor?.up) {
        new Notification(`UP: ${monitor.friendlyName}`, {
          body: `"${monitor.friendlyName}" has recovered. Congrats!`,
          icon: "/logo.svg",
        })
      }
    }
  }, [monitors])

  const renderedMonitors = monitors.map((val) => <MonitorCard monitorData={val} key={val.name} />)

  return (
    <div className="h-full p-4">
      <h1 className="font-bold text-3xl">Monitors</h1>
      <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2 py-2">{renderedMonitors}</div>
    </div>
  )
}
