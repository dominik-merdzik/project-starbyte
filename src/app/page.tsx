// app/page.tsx
import Image from "next/image";

export default function Home() {
  return (
    <div
      className="min-h-screen bg-[#212121] text-gray-100 flex flex-col"
      
    >
      {/* Hero Section */}
      <section className="relative w-full h-[95vh] bg-[#2c2c2c] flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-5xl font-bold">Project Starbyte</h1>
          <p className="mt-4 text-xl text-[#adadad]">
            Embark on an interstellar adventure
          </p>
          <p className="mt-4 w-50 text-xl">
            ˚　　　　✦　　　.　　. 　 ˚　.　　　　　 . ✦　　　 　˚　　　　 . ★⋆. ࿐࿔
            　　.   　　˚　　 　　*　　 　　✦　　　.　　.　　　✦　˚ 　　　　 ˚　.˚　　　　
            ✦　　　.　　. 　 ˚　.　　　　 　　 　　　　    	ੈ✧̣̇˳·˖✶   ✦　　
          </p>
        </div>
      </section>

      {/* Demo Section */}
      <section className="w-full h-screen flex flex-col items-center justify-center px-4">
        <h2 className="text-4xl font-bold text-center mb-6">Download the Latest Demo</h2>
        <p className="text-lg mb-8 text-center max-w-xl">
          Visit our GitHub repository to download the latest executable.
          Stay updated with all the newest features and improvements!
        </p>
        <a
          href="https://github.com/dominik-merdzik/project-starbyte"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center space-x-2 bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-full transition"
        >
          <Image
            src="/github-mark-white.svg"
            alt="GitHub Logo"
            width={24}
            height={24}
          />
          <span>Go to GitHub</span>
        </a>
      </section>

      {/* Footer */}
      <footer className="py-4 text-center">
        <p>&copy; {new Date().getFullYear()} Project Starbyte. All rights reserved.</p>
      </footer>
    </div>
  );
}
