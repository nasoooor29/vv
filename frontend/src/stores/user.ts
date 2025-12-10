import { T } from "@/types";
import { atomWithStorage } from "jotai/utils";

// Set the string key and the initial value
export const userAtom = atomWithStorage<{
  user: T.User | null;
  session: T.UserSession | null;
}>("user", {
  user: null,
  session: null,
});
// const Page = () => {
//   // Consume persisted state like any other atom
//   const [darkMode, setDarkMode] = useAtom(darkModeAtom)
//   const toggleDarkMode = () => setDarkMode(!darkMode)
//   return (
//     <>
//       <h1>Welcome to {darkMode ? 'dark' : 'light'} mode!</h1>
//       <button onClick={toggleDarkMode}>toggle theme</button>
//     </>
//   )
// }
