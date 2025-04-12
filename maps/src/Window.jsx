import { useEffect ,useState} from "react";

const useWindowSize = () => {
    const [size, setSize] = useState([window.innerWidth, window.innerHeight]);

    useEffect(() => {
        const handleResize = () => {
            setSize([window.innerWidth, window.innerHeight]);
        };

        window.addEventListener('resize', handleResize);
        screen.orientation.addEventListener("change",handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
            screen.orientation.removeEventListener('change', handleResize);
        };
    },[]);

    const [width, height] = size;
    return {width,height};
}

export default useWindowSize;