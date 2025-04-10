import { useEffect } from "react";

import svg from "./assets/crosshair.svg"

const SVGButton = (props) => {

    useEffect( () => {
    },[])

    return (<>
      <button {...props} className="svgBtn">
        <object type="image/svg+xml" data={svg} width={props.size} height={props.size}/>
      </button>
    </>)
}
export default SVGButton;