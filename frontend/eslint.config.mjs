const config = [
  ...nextCoreWebVitals,
  ...nextTypeScript,
  {
    ignores: ["./next/**", "node_modules?**", "next-env.d.ts"],
  },
];

export default config;
