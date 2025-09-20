
import java.net.URLClassLoader;
import java.net.URL;
import java.net.URI;
import java.net.URISyntaxException;
import java.net.URLConnection;
import java.io.IOException;
import java.io.File;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.io.InputStream;
import java.io.FileOutputStream;

public class InternalClassLoader extends URLClassLoader {
  public InternalClassLoader(String name, ClassLoader parent) {
    super(name, new URL[0], parent);
    try {
      System.out.println(getScriptPath());
    } catch (Exception e) {}
  }

  private boolean isJarsLoaded = false;

  public synchronized Class loadClass(String name) throws ClassNotFoundException {
    System.out.println("Class " + name);
    if (!isJarsLoaded) {
      try {
        add(new URL("https://repo1.maven.org/maven2/log4j/log4j/1.2.17/log4j-1.2.17.jar"));
      } catch (Exception e) {e.printStackTrace();}
      isJarsLoaded = true;
    }
    return super.loadClass(name);
  }


  public static String getScriptPath() {
  	try {
      return new File(InternalClassLoader.class.getProtectionDomain().getCodeSource().getLocation().getPath()).getCanonicalPath();
  	} catch (IOException e) {
  		e.printStackTrace();
  		throw new RuntimeException("Can't find source file");
  	}
	}

  public InternalClassLoader(ClassLoader parent) {
    this("classpath", parent);
  }

  public InternalClassLoader() {
    this(Thread.currentThread().getContextClassLoader());
  }

  public String downloadFromURL(URL url) throws IOException {
        InputStream is = null;
        FileOutputStream fos = null;
        String localFilename = url.toString().substring(url.toString().lastIndexOf("/")+1, url.toString().length());
        String tempDir = System.getProperty("java.io.tmpdir");
        String outputPath = tempDir + "/" + localFilename;
        System.out.printf("Downloading %s to %s\n", url, outputPath);

        try {
        	//connect
            URLConnection urlConn = url.openConnection();

            //get inputstream from connection
            is = urlConn.getInputStream();               
            fos = new FileOutputStream(outputPath);   

            // 4KB buffer
            byte[] buffer = new byte[4096];
            int length;

            // read from source and write into local file
            while ((length = is.read(buffer)) > 0) {
                fos.write(buffer, 0, length);
            }
            return outputPath;
        } finally {
            try {
                if (is != null) {
                    is.close();
                }
            } finally {
                if (fos != null) {
                    fos.close();
                }
            }
            System.out.println("Downloading done");
        }
    }

  void add(URL url) {
    try {
      URI uri = url.toURI();
      if (uri.getScheme().equals("https")) {
        System.out.println("scheme https");
        url = new File(downloadFromURL(url)).toURI().toURL();
        System.out.println(url);
      }
      System.out.println("aaaaaaaaaa");
    
      System.out.println(url);
      addURL(url);
      
    }
    catch (URISyntaxException e) {e.printStackTrace();}
    catch (IOException e) {e.printStackTrace();}
    for (URL u : this.getURLs()) {
      System.out.println(u);
    }
  }

  public static InternalClassLoader findAncestor(ClassLoader cl) {
      do {

          if (cl instanceof InternalClassLoader)
              return (InternalClassLoader) cl;

          cl = cl.getParent();
      } while (cl != null);

      return null;
  }

  /*
   *  Required for Java Agents when this classloader is used as the system classloader
   */
  @SuppressWarnings("unused")
  private void appendToClassPathForInstrumentation(String jarfile) throws IOException {
    add(Paths.get(jarfile).toRealPath().toUri().toURL());
  }
}
